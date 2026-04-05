import AppKit
import SwiftUI

struct ItemDetailView: View {
    @EnvironmentObject var vault: VaultClient

    let item: VaultItem
    let onEdit: () -> Void

    @State private var showingVersions = false
    @State private var versions: [ItemVersion] = []
    @State private var versionsError: String?
    @State private var revealedFields: Set<String> = []
    @State private var showCopiedToast = false
    @State private var copiedFieldName = ""
    @State private var showDeleteConfirmation = false
    @State private var copiedFieldKey: String?

    private var sensitiveKeys: Set<String> {
        item.type.sensitiveFieldKeys
    }

    private var orderedFieldKeys: [String] {
        let priority = [
            "username", "email", "user", "login",
            "password", "api_key", "api_secret", "secret",
            "private_key", "public_key", "passphrase",
            "url", "endpoint",
            "card_number", "cardholder", "expiry", "cvv",
            "full_name", "phone", "address",
            "credential_id", "relying_party", "user_handle",
        ]

        let allKeys = item.fields.keys.sorted()
        var ordered: [String] = []

        for key in priority where allKeys.contains(key) {
            ordered.append(key)
        }

        for key in allKeys where !ordered.contains(key) {
            ordered.append(key)
        }

        return ordered
    }

    private var preferredPrimaryKeys: [String] {
        switch item.type {
        case .login:
            ["username", "email", "password", "url"]
        case .apikey:
            ["api_key", "api_secret", "endpoint"]
        case .sshkey:
            ["private_key", "public_key", "passphrase"]
        case .note:
            []
        case .creditcard:
            ["cardholder", "card_number", "expiry", "cvv"]
        case .identity:
            ["full_name", "email", "phone", "address"]
        case .passkey:
            ["relying_party", "credential_id", "user_handle"]
        case .custom:
            []
        }
    }

    private var primaryFieldKeys: [String] {
        orderedFieldKeys.filter {
            preferredPrimaryKeys.contains($0) && !fieldValue(for: $0).isEmpty
        }
    }

    private var additionalFieldKeys: [String] {
        orderedFieldKeys.filter {
            !primaryFieldKeys.contains($0) && !fieldValue(for: $0).isEmpty
        }
    }

    private var primaryIdentityField: (key: String, value: String)? {
        item.preferredIdentityField
    }

    private var primarySecretField: (key: String, value: String)? {
        item.preferredSecretField
    }

    private var primaryURLField: (key: String, value: String)? {
        item.preferredURLField
    }

    private var hasQuickActions: Bool {
        primaryIdentityField != nil || primarySecretField != nil || primaryURLField != nil
    }

    private var summarySubtitle: String? {
        primaryIdentityField?.value ?? primaryURLField?.value
    }

    private var primarySectionTitle: String {
        switch item.type {
        case .identity:
            "Contact"
        case .creditcard:
            "Card"
        case .passkey:
            "Passkey"
        case .note, .custom:
            "Details"
        default:
            "Credentials"
        }
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: ZPTheme.spacing16) {
                summaryView

                if hasQuickActions {
                    detailSection("Quick Actions") {
                        quickActionsView
                    }
                }

                if !primaryFieldKeys.isEmpty {
                    fieldSection(title: primarySectionTitle, keys: primaryFieldKeys)
                }

                if !additionalFieldKeys.isEmpty {
                    fieldSection(title: "Additional Fields", keys: additionalFieldKeys)
                }

                if !item.notes.isEmpty {
                    detailSection("Notes") {
                        Text(item.notes)
                            .font(.body)
                            .foregroundStyle(ZPTheme.textPrimary)
                            .textSelection(.enabled)
                            .frame(maxWidth: .infinity, alignment: .leading)
                    }
                }

                if !item.tags.isEmpty {
                    detailSection("Tags") {
                        FlowLayout(spacing: 8) {
                            ForEach(item.tags, id: \.self) { tag in
                                Text(tag)
                                    .font(.caption.weight(.semibold))
                                    .foregroundStyle(ZPTheme.textSecondary)
                                    .padding(.horizontal, ZPTheme.spacing10)
                                    .padding(.vertical, ZPTheme.spacing6)
                                    .background(ZPTheme.chipBackground, in: Capsule())
                                    .overlay(
                                        Capsule()
                                            .stroke(ZPTheme.panelBorder, lineWidth: 1)
                                    )
                            }
                        }
                        .frame(maxWidth: .infinity, alignment: .leading)
                    }
                }

                detailSection("Details") {
                    detailMetaRow(title: "Modified", value: item.updatedAt.formatted(date: .abbreviated, time: .shortened))
                    Divider()
                    detailMetaRow(title: "Created", value: item.createdAt.formatted(date: .abbreviated, time: .shortened))
                    Divider()
                    detailMetaRow(title: "Version", value: "v\(item.version)")
                    Divider()
                    Button {
                        loadVersionsAndShow()
                    } label: {
                        Label("Version History", systemImage: "clock.arrow.circlepath")
                    }
                    .buttonStyle(.bordered)
                }
            }
            .padding(ZPTheme.spacing24)
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .background(ZPTheme.workspaceBackground)
        .toolbar {
            ToolbarItemGroup {
                Button {
                    onEdit()
                } label: {
                    Label("Edit", systemImage: "pencil")
                }
                .help("Edit item")

                Menu {
                    Button {
                        loadVersionsAndShow()
                    } label: {
                        Label("Version History", systemImage: "clock.arrow.circlepath")
                    }

                    Divider()

                    Button(role: .destructive) {
                        showDeleteConfirmation = true
                    } label: {
                        Label("Delete…", systemImage: "trash")
                    }
                } label: {
                    Image(systemName: "ellipsis.circle")
                        .accessibilityLabel("More item actions")
                }
                .help("More item actions")
            }
        }
        .toast(isShowing: $showCopiedToast, message: "Copied \(copiedFieldName)")
        .alert("Delete Item", isPresented: $showDeleteConfirmation) {
            Button("Delete", role: .destructive) {
                Task {
                    do { try await vault.deleteItem(id: item.id) }
                    catch { vault.lastError = error.localizedDescription }
                }
            }
            Button("Cancel", role: .cancel) {}
        } message: {
            Text("Are you sure you want to delete \"\(item.name)\"? This action cannot be undone.")
        }
        .sheet(isPresented: $showingVersions) {
            VersionHistoryView(
                itemID: item.id,
                versions: versions,
                error: versionsError
            ) { version in
                Task {
                    do {
                        try await vault.restoreVersion(itemID: item.id, version: version)
                        showingVersions = false
                    } catch {
                        versionsError = error.localizedDescription
                    }
                }
            }
        }
    }

    private var summaryView: some View {
        HStack(alignment: .top, spacing: ZPTheme.spacing16) {
            ZStack {
                RoundedRectangle(cornerRadius: ZPTheme.radiusLarge, style: .continuous)
                    .fill(item.type.color.opacity(0.14))

                Image(systemName: item.type.symbolName)
                    .font(.title2.weight(.semibold))
                    .foregroundStyle(item.type.color)
            }
            .frame(width: 48, height: 48)

            VStack(alignment: .leading, spacing: ZPTheme.spacing6) {
                Text(item.name.isEmpty ? "(Untitled)" : item.name)
                    .font(.title2.weight(.semibold))
                    .foregroundStyle(ZPTheme.textPrimary)

                HStack(spacing: ZPTheme.spacing8) {
                    Text(item.type.displayName)
                        .font(.callout.weight(.semibold))
                        .foregroundStyle(ZPTheme.accent)
                        .padding(.horizontal, ZPTheme.spacing10)
                        .padding(.vertical, ZPTheme.spacing6)
                        .background(ZPTheme.pillBackground, in: Capsule())
                        .overlay(
                            Capsule()
                                .stroke(ZPTheme.pillBorder, lineWidth: 1)
                        )

                    if item.favorite {
                        Label("Favorite", systemImage: "star.fill")
                            .font(.caption.weight(.semibold))
                            .foregroundStyle(.yellow)
                    }
                }

                if let summarySubtitle {
                    Text(summarySubtitle)
                        .font(.callout)
                        .foregroundStyle(ZPTheme.textSecondary)
                        .lineLimit(1)
                        .textSelection(.enabled)
                }
            }

            Spacer()
        }
        .padding(ZPTheme.spacing20)
        .zpSurface(.elevated)
    }

    private var quickActionsView: some View {
        FlowLayout(spacing: 8) {
            if let identity = primaryIdentityField {
                quickActionButton(
                    title: "Copy \(fieldTitle(for: identity.key))",
                    systemImage: "doc.on.doc"
                ) {
                    copyField(key: identity.key, value: identity.value)
                }
            }

            if let secret = primarySecretField {
                quickActionButton(
                    title: "Copy \(fieldTitle(for: secret.key))",
                    systemImage: "key.fill"
                ) {
                    copyField(key: secret.key, value: secret.value)
                }
            }

            if primaryURLField != nil {
                quickActionButton(
                    title: primaryURLActionTitle,
                    systemImage: "globe"
                ) {
                    openPrimaryURL()
                }
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.vertical, 2)
    }

    @ViewBuilder
    private func fieldRow(key: String, value: String) -> some View {
        let isSensitive = sensitiveKeys.contains(key)
        let isRevealed = revealedFields.contains(key)
        let isCopied = copiedFieldKey == key
        let title = fieldTitle(for: key)

        VStack(alignment: .leading, spacing: ZPTheme.spacing8) {
            Text(title.uppercased())
                .font(.caption.weight(.bold))
                .foregroundStyle(ZPTheme.textTertiary)

            HStack(spacing: ZPTheme.spacing12) {
                Group {
                    if isSensitive && !isRevealed {
                        Text("●●●●●●●●●●●●")
                            .font(.system(.body, design: .monospaced))
                            .foregroundStyle(ZPTheme.textSecondary)
                            .accessibilityLabel(title)
                            .accessibilityValue("Hidden")
                    } else {
                        Text(value)
                            .font(isSensitive ? .system(.body, design: .monospaced) : .body)
                            .foregroundStyle(ZPTheme.textPrimary)
                            .textSelection(.enabled)
                            .accessibilityLabel(title)
                            .accessibilityValue(value)
                    }
                }
                .frame(maxWidth: .infinity, alignment: .leading)

                HStack(spacing: 4) {
                    if isSensitive {
                        Button {
                            if isRevealed {
                                revealedFields.remove(key)
                            } else {
                                revealedFields.insert(key)
                            }
                        } label: {
                            Image(systemName: isRevealed ? "eye.slash" : "eye")
                        }
                        .buttonStyle(.borderless)
                        .help(isRevealed ? "Hide" : "Reveal")
                        .accessibilityLabel(isRevealed ? "Hide \(title)" : "Reveal \(title)")
                    }

                    Button {
                        copyField(key: key, value: value)
                    } label: {
                        Image(systemName: isCopied ? "checkmark.circle.fill" : "doc.on.doc")
                            .foregroundStyle(isCopied ? ZPTheme.success : ZPTheme.textSecondary)
                    }
                    .buttonStyle(.borderless)
                    .help("Copy \(title)")
                    .accessibilityLabel("Copy \(title)")
                    .animation(.easeInOut(duration: 0.2), value: isCopied)
                }
            }
        }
        .padding(.vertical, ZPTheme.spacing10)
    }

    @ViewBuilder
    private func quickActionButton(
        title: String,
        systemImage: String,
        action: @escaping () -> Void
    ) -> some View {
        Button(action: action) {
            Label(title, systemImage: systemImage)
        }
        .buttonStyle(.bordered)
        .controlSize(.small)
    }

    private func copyField(key: String, value: String) {
        let secs = vault.clipboardAutoClearEnabled ? vault.clipboardAutoClearSeconds : 0
        ClipboardService.shared.copySensitive(value, clearAfterSeconds: secs)
        copiedFieldName = fieldTitle(for: key)

        withAnimation {
            copiedFieldKey = key
            showCopiedToast = true
        }

        Task {
            try? await Task.sleep(nanoseconds: 1_500_000_000)
            withAnimation {
                copiedFieldKey = nil
            }
        }
    }

    private var primaryURLActionTitle: String {
        primaryURLField?.key == "endpoint" ? "Open Endpoint" : "Open Website"
    }

    private func fieldTitle(for key: String) -> String {
        key.replacingOccurrences(of: "_", with: " ").capitalized
    }

    private func fieldValue(for key: String) -> String {
        item.fields[key, default: ""].trimmingCharacters(in: .whitespacesAndNewlines)
    }

    private func firstNonEmptyField(in keys: [String]) -> (key: String, value: String)? {
        for key in keys {
            let value = fieldValue(for: key)
            if !value.isEmpty {
                return (key, value)
            }
        }
        return nil
    }

    private func openPrimaryURL() {
        guard let field = primaryURLField else { return }
        let value = field.value
        let normalizedValue = value.hasPrefix("http://") || value.hasPrefix("https://")
            ? value
            : "https://\(value)"

        guard let url = URL(string: normalizedValue) else { return }
        NSWorkspace.shared.open(url)
    }

    @ViewBuilder
    private func detailSection<Content: View>(_ title: String, @ViewBuilder content: () -> Content) -> some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing14) {
            Text(title)
                .font(.headline)
                .foregroundStyle(ZPTheme.textPrimary)

            content()
        }
        .padding(ZPTheme.spacing18)
        .frame(maxWidth: .infinity, alignment: .leading)
        .zpSurface(.card)
    }

    @ViewBuilder
    private func fieldSection(title: String, keys: [String]) -> some View {
        detailSection(title) {
            VStack(spacing: 0) {
                ForEach(Array(keys.enumerated()), id: \.element) { index, key in
                    fieldRow(key: key, value: item.fields[key] ?? "")

                    if index < keys.count - 1 {
                        Divider()
                    }
                }
            }
        }
    }

    @ViewBuilder
    private func detailMetaRow(title: String, value: String) -> some View {
        HStack {
            Text(title)
                .foregroundStyle(ZPTheme.textSecondary)

            Spacer()

            Text(value)
                .foregroundStyle(ZPTheme.textPrimary)
        }
    }

    private func loadVersionsAndShow() {
        versionsError = nil
        Task {
            do {
                versions = try await vault.versionHistory(itemID: item.id)
                showingVersions = true
            } catch {
                versionsError = error.localizedDescription
                showingVersions = true
            }
        }
    }
}

// MARK: - Flow Layout

struct FlowLayout: Layout {
    var spacing: CGFloat = 6

    func sizeThatFits(proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) -> CGSize {
        let result = arrange(proposal: proposal, subviews: subviews)
        return result.size
    }

    func placeSubviews(in bounds: CGRect, proposal: ProposedViewSize, subviews: Subviews, cache: inout ()) {
        let result = arrange(proposal: proposal, subviews: subviews)
        for (index, position) in result.positions.enumerated() {
            subviews[index].place(
                at: CGPoint(x: bounds.minX + position.x, y: bounds.minY + position.y),
                proposal: .unspecified
            )
        }
    }

    private func arrange(proposal: ProposedViewSize, subviews: Subviews) -> (size: CGSize, positions: [CGPoint]) {
        let maxWidth = proposal.width ?? .infinity
        var positions: [CGPoint] = []
        var x: CGFloat = 0
        var y: CGFloat = 0
        var rowHeight: CGFloat = 0
        var maxX: CGFloat = 0

        for subview in subviews {
            let size = subview.sizeThatFits(.unspecified)
            if x + size.width > maxWidth, x > 0 {
                x = 0
                y += rowHeight + spacing
                rowHeight = 0
            }
            positions.append(CGPoint(x: x, y: y))
            rowHeight = max(rowHeight, size.height)
            x += size.width + spacing
            maxX = max(maxX, x)
        }

        return (CGSize(width: maxX, height: y + rowHeight), positions)
    }
}

// MARK: - Version History

private struct VersionHistoryView: View {
    @Environment(\.dismiss) private var dismiss

    let itemID: String
    let versions: [ItemVersion]
    let error: String?
    let onRestore: (Int) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Version History")
                .font(.title2)
                .bold()

            if let error {
                Text(error)
                    .foregroundStyle(.red)
                    .textSelection(.enabled)
            }

            List {
                ForEach(versions) { v in
                    HStack {
                        Label("v\(v.version)", systemImage: "clock")
                        Spacer()
                        Text(v.savedAt.formatted(date: .abbreviated, time: .shortened))
                            .foregroundStyle(.secondary)
                        Button("Restore") { onRestore(v.version) }
                    }
                }
            }

            HStack {
                Spacer()
                Button("Close") { dismiss() }
                    .keyboardShortcut(.cancelAction)
            }
        }
        .padding(20)
        .frame(width: 520, height: 420)
    }
}
