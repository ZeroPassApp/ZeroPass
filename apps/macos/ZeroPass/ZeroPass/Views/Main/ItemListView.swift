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

struct ItemListView: View {
    @EnvironmentObject var vault: VaultClient

    let category: SidebarCategory?
    @Binding var selectedItemID: String?
    @Binding var editingItem: VaultItem?

    @State private var sortOption: ItemSortOption = .name
    @State private var sortAscending = true
    @State private var searchText = ""
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
            items = items.filter {
                $0.name.localizedCaseInsensitiveContains(searchText) ||
                $0.notes.localizedCaseInsensitiveContains(searchText) ||
                $0.fields.filter({ !$0.value.isEmpty && !$0.key.isEmpty &&
                    !["password", "secret", "private_key", "api_key", "secret_key", "cvv", "pin"].contains($0.key) })
                    .values.contains(where: { $0.localizedCaseInsensitiveContains(searchText) })
            }
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
                            .accessibilityHint("Double-click to view details")
                            .contextMenu {
                                contextMenuItems(for: item)
                            }
                    }
                }
            }
        }
        .navigationTitle(categoryTitle)
        .searchable(text: $searchText, placement: .sidebar, prompt: "Search items")
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
                    Image(systemName: "arrow.up.arrow.down")
                }
                .help("Sort items")
                .accessibilityLabel("Sort items")
            }
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
                if case .all = category {
                    Button("Add Your First Item") {
                        // Will be handled by parent
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
        Button {
            copyUsername(item)
        } label: {
            Label("Copy Username", systemImage: "person")
        }

        Button {
            copySecret(item)
        } label: {
            Label("Copy Password", systemImage: "key")
        }

        if let url = item.fields["url"], !url.isEmpty {
            Button {
                if let u = URL(string: url.hasPrefix("http") ? url : "https://\(url)") {
                    NSWorkspace.shared.open(u)
                }
            } label: {
                Label("Open URL", systemImage: "globe")
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

    private func copyUsername(_ item: VaultItem) {
        let candidates = ["username", "email", "user", "login", "cardholder", "full_name"]
        guard let value = candidates.compactMap({ item.fields[$0] }).first(where: { !$0.isEmpty }) else { return }
        let secs = vault.clipboardAutoClearEnabled ? vault.clipboardAutoClearSeconds : 0
        ClipboardService.shared.copySensitive(value, clearAfterSeconds: secs)
    }

    private func copySecret(_ item: VaultItem) {
        let candidates = ["password", "api_secret", "secret", "api_key", "private_key", "cvv"]
        guard let value = candidates.compactMap({ item.fields[$0] }).first(where: { !$0.isEmpty }) else { return }
        let secs = vault.clipboardAutoClearEnabled ? vault.clipboardAutoClearSeconds : 0
        ClipboardService.shared.copySensitive(value, clearAfterSeconds: secs)
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
        if !item.notes.isEmpty { return String(item.notes.prefix(50)) }
        return ""
    }

    var body: some View {
        HStack(spacing: 8) {
            Image(systemName: item.type.symbolName)
                .foregroundStyle(item.type.color)
                .font(.system(size: 14))
                .frame(width: 20)

            VStack(alignment: .leading, spacing: 2) {
                Text(item.name.isEmpty ? "(Untitled)" : item.name)
                    .font(.system(size: 13, weight: .medium))
                    .lineLimit(1)

                if !subtitle.isEmpty {
                    Text(subtitle)
                        .font(.system(size: 11))
                        .foregroundStyle(.secondary)
                        .lineLimit(1)
                }
            }

            Spacer()

            if item.favorite {
                Image(systemName: "star.fill")
                    .foregroundStyle(.yellow)
                    .font(.system(size: 10))
            }
        }
        .padding(.vertical, 2)
    }
}
