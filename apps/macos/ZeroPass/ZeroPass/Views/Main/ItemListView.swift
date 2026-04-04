import SwiftUI

enum ItemSortOption: String, CaseIterable {
    case name = "Name"
    case updatedAt = "Date Modified"
    case createdAt = "Date Created"
    case type = "Type"

    var sortKey: String {
        switch self {
        case .name: return "name"
        case .updatedAt: return "updated_at"
        case .createdAt: return "created_at"
        case .type: return "type"
        }
    }
}

enum VaultItemSearchMatcher {
    static func matches(item: VaultItem, query: String) -> Bool {
        let normalizedQuery = query.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !normalizedQuery.isEmpty else { return true }

        if item.name.localizedCaseInsensitiveContains(normalizedQuery) ||
            item.notes.localizedCaseInsensitiveContains(normalizedQuery) {
            return true
        }

        return searchableFieldValues(for: item).contains {
            $0.localizedCaseInsensitiveContains(normalizedQuery)
        }
    }

    private static func searchableFieldValues(for item: VaultItem) -> [String] {
        item.fields.compactMap { key, value in
            let normalizedValue = value.trimmingCharacters(in: .whitespacesAndNewlines)
            guard !key.isEmpty, !normalizedValue.isEmpty else { return nil }
            guard !item.type.sensitiveFieldKeys.contains(key) else { return nil }
            return normalizedValue
        }
    }
}

struct ItemListView: View {
    @EnvironmentObject var vault: VaultClient

    let category: SidebarCategory?
    @Binding var searchText: String
    @Binding var selectedItemID: String?
    @Binding var editingItem: VaultItem?
    let onCreateItem: () -> Void

    @State private var sortOption: ItemSortOption = .name
    @State private var sortAscending = true
    @State private var itemToDelete: VaultItem?

    private var filteredItems: [VaultItem] {
        var items: [VaultItem]

        switch category {
        case .all, .none:
            items = vault.filteredItems(sortBy: sortOption.sortKey, sortOrder: sortAscending ? "asc" : "desc")
        case .favorites:
            items = vault.filteredItems(favoritesOnly: true, sortBy: sortOption.sortKey, sortOrder: sortAscending ? "asc" : "desc")
        case .type(let type):
            items = vault.filteredItems(type: type, sortBy: sortOption.sortKey, sortOrder: sortAscending ? "asc" : "desc")
        case .tag(let tag):
            items = vault.filteredItems(tag: tag, sortBy: sortOption.sortKey, sortOrder: sortAscending ? "asc" : "desc")
        }

        if !searchText.isEmpty {
            items = items.filter { VaultItemSearchMatcher.matches(item: $0, query: searchText) }
        }

        return items
    }

    private var categoryTitle: String {
        switch category {
        case .all, .none: return "All Items"
        case .favorites: return "Favorites"
        case .type(let type): return type.displayName
        case .tag(let tag): return tag
        }
    }

    private var filteredItemIDs: [String] {
        filteredItems.map(\.id)
    }

    var body: some View {
        Group {
            if filteredItems.isEmpty {
                emptyState
            } else {
                List(selection: $selectedItemID) {
                    ForEach(filteredItems) { item in
                        ItemRow(item: item)
                            .tag(item.id)
                            .accessibilityLabel("\(item.name), \(item.type.displayName)")
                            .accessibilityHint("Select to view details")
                            .contextMenu {
                                contextMenuItems(for: item)
                            }
                    }
                }
            }
        }
        .navigationTitle(categoryTitle)
        .toolbar {
            ToolbarItem(placement: .automatic) {
                Menu {
                    ForEach(ItemSortOption.allCases, id: \.self) { option in
                        Button {
                            if sortOption == option {
                                sortAscending.toggle()
                            } else {
                                sortOption = option
                                sortAscending = true
                            }
                        } label: {
                            HStack {
                                Text(option.rawValue)
                                if sortOption == option {
                                    Image(systemName: sortAscending ? "chevron.up" : "chevron.down")
                                }
                            }
                        }
                    }
                } label: {
                    Label("Sort", systemImage: "arrow.up.arrow.down")
                }
                .help("Sort items")
                .accessibilityLabel("Sort items")
            }
        }
        .onAppear {
            ensureVisibleSelection(in: filteredItemIDs)
        }
        .onChange(of: filteredItemIDs) { _, ids in
            ensureVisibleSelection(in: ids)
        }
        .alert("Delete Item", isPresented: Binding(
            get: { itemToDelete != nil },
            set: { if !$0 { itemToDelete = nil } }
        )) {
            Button("Delete", role: .destructive) {
                if let item = itemToDelete {
                    Task {
                        do { try await vault.deleteItem(id: item.id) }
                        catch { vault.lastError = error.localizedDescription }
                    }
                }
            }
            Button("Cancel", role: .cancel) { itemToDelete = nil }
        } message: {
            Text("Are you sure you want to delete \"\(itemToDelete?.name ?? "")\"? This action cannot be undone.")
        }
    }

    @ViewBuilder
    private var emptyState: some View {
        if searchText.isEmpty {
            ContentUnavailableView {
                Label(emptyTitle, systemImage: emptyIcon)
            } description: {
                Text(emptyDescription)
            } actions: {
                if category == .all || category == nil {
                    Button("Add Your First Item") {
                        onCreateItem()
                    }
                }
            }
        } else {
            ContentUnavailableView.search(text: searchText)
        }
    }

    private var emptyTitle: String {
        switch category {
        case .favorites: return "No Favorites"
        case .type(let type): return "No \(type.displayName) Items"
        case .tag(let tag): return "No Items Tagged \"\(tag)\""
        default: return "Vault is Empty"
        }
    }

    private var emptyIcon: String {
        switch category {
        case .favorites: return "star"
        case .type(let type): return type.symbolName
        case .tag: return "tag"
        default: return "tray"
        }
    }

    private var emptyDescription: String {
        switch category {
        case .favorites: return "Star items to see them here."
        case .type: return "Items of this type will appear here."
        case .tag: return "Tag items to organize them."
        default: return "Create your first item to get started."
        }
    }

    @ViewBuilder
    private func contextMenuItems(for item: VaultItem) -> some View {
        let identityField = copyableIdentityField(for: item)
        let secretField = copyableSecretField(for: item)
        let linkField = primaryURLField(for: item)

        Button {
            if let identityField {
                copyFieldValue(identityField.value)
            }
        } label: {
            Label(identityField.map { "Copy \(fieldTitle(for: $0.key))" } ?? "Copy Identifier", systemImage: "doc.on.doc")
        }
        .disabled(identityField == nil)

        Button {
            if let secretField {
                copyFieldValue(secretField.value)
            }
        } label: {
            Label(secretField.map { "Copy \(fieldTitle(for: $0.key))" } ?? "Copy Secret", systemImage: "key")
        }
        .disabled(secretField == nil)

        if let linkField {
            Button {
                openURL(linkField.value)
            } label: {
                Label(linkField.key == "endpoint" ? "Open Endpoint" : "Open Website", systemImage: "globe")
            }
        }

        Divider()

        Button {
            Task { try? await vault.toggleFavorite(item: item) }
        } label: {
            Label(item.favorite ? "Remove from Favorites" : "Add to Favorites",
                  systemImage: item.favorite ? "star.slash" : "star")
        }

        Button {
            editingItem = item
        } label: {
            Label("Edit", systemImage: "pencil")
        }

        Divider()

        Button(role: .destructive) {
            itemToDelete = item
        } label: {
            Label("Delete", systemImage: "trash")
        }
    }

    private func copyableIdentityField(for item: VaultItem) -> (key: String, value: String)? {
        firstNonEmptyField(in: item, keys: ["username", "email", "user", "login", "cardholder", "full_name", "relying_party"])
    }

    private func copyableSecretField(for item: VaultItem) -> (key: String, value: String)? {
        firstNonEmptyField(in: item, keys: ["password", "api_secret", "secret", "api_key", "private_key", "cvv", "card_number", "credential_id", "passphrase"])
    }

    private func primaryURLField(for item: VaultItem) -> (key: String, value: String)? {
        firstNonEmptyField(in: item, keys: ["url", "endpoint"])
    }

    private func firstNonEmptyField(in item: VaultItem, keys: [String]) -> (key: String, value: String)? {
        for key in keys {
            let value = item.fields[key, default: ""].trimmingCharacters(in: .whitespacesAndNewlines)
            if !value.isEmpty {
                return (key, value)
            }
        }
        return nil
    }

    private func copyFieldValue(_ value: String) {
        let secs = vault.clipboardAutoClearEnabled ? vault.clipboardAutoClearSeconds : 0
        ClipboardService.shared.copySensitive(value, clearAfterSeconds: secs)
    }

    private func openURL(_ value: String) {
        let normalizedValue = value.hasPrefix("http://") || value.hasPrefix("https://")
            ? value
            : "https://\(value)"

        if let url = URL(string: normalizedValue) {
            NSWorkspace.shared.open(url)
        }
    }

    private func fieldTitle(for key: String) -> String {
        key.replacingOccurrences(of: "_", with: " ").capitalized
    }

    private func ensureVisibleSelection(in ids: [String]) {
        guard !ids.isEmpty else {
            selectedItemID = nil
            return
        }

        if let selectedItemID, ids.contains(selectedItemID) {
            return
        }

        selectedItemID = ids.first
    }
}

// MARK: - Item Row

private struct ItemRow: View {
    let item: VaultItem

    private var subtitle: String {
        if let username = item.fields["username"], !username.isEmpty { return username }
        if let email = item.fields["email"], !email.isEmpty { return email }
        if let url = item.fields["url"], !url.isEmpty { return url }
        if let endpoint = item.fields["endpoint"], !endpoint.isEmpty { return endpoint }
        return ""
    }

    var body: some View {
        HStack(spacing: 8) {
            Image(systemName: item.type.symbolName)
                .font(.body)
                .frame(width: 20)

            VStack(alignment: .leading, spacing: 2) {
                Text(item.name.isEmpty ? "(Untitled)" : item.name)
                    .font(.body.weight(.medium))
                    .lineLimit(1)

                if !subtitle.isEmpty {
                    Text(subtitle)
                        .font(.callout)
                        .foregroundStyle(.secondary)
                        .lineLimit(1)
                }
            }

            Spacer()

            if item.favorite {
                Image(systemName: "star.fill")
                    .foregroundStyle(.yellow)
                    .font(.caption)
            }
        }
        .padding(.vertical, 4)
    }
}
