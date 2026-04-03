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

    var body: some View {
        Form {
            // Header
            Section {
                HStack(spacing: 10) {
                    Image(systemName: item.type.symbolName)
                        .font(.title2)
                        .foregroundStyle(item.type.color)

                    VStack(alignment: .leading, spacing: 2) {
                        Text(item.name.isEmpty ? "(Untitled)" : item.name)
                            .font(.title3)
                            .bold()

                        Text(item.type.displayName)
                            .font(.callout)
                            .foregroundStyle(.secondary)
                    }

                    Spacer()

                    if item.favorite {
                        Image(systemName: "star.fill")
                            .foregroundStyle(.yellow)
                            .accessibilityLabel("Favorite")
                    }
                }
            }

            // Fields
            if !item.fields.isEmpty {
                Section("Fields") {
                    ForEach(orderedFieldKeys, id: \.self) { key in
                        fieldRow(key: key, value: item.fields[key] ?? "")
                    }
                }
            }

            // Notes
            if !item.notes.isEmpty {
                Section("Notes") {
                    Text(item.notes)
                        .font(.body)
                        .textSelection(.enabled)
                        .frame(maxWidth: .infinity, alignment: .leading)
                }
            }

            // Tags
            if !item.tags.isEmpty {
                Section("Tags") {
                    FlowLayout(spacing: 6) {
                        ForEach(item.tags, id: \.self) { tag in
                            Text(tag)
                                .font(.caption)
                                .padding(.horizontal, 8)
                                .padding(.vertical, 3)
                                .background(.quaternary)
                                .clipShape(Capsule())
                        }
                    }
                }
            }

            // Metadata
            Section("Info") {
                LabeledContent("Created") {
                    Text(item.createdAt.formatted(date: .abbreviated, time: .shortened))
                        .foregroundStyle(.secondary)
                }
                LabeledContent("Modified") {
                    Text(item.updatedAt.formatted(date: .abbreviated, time: .shortened))
                        .foregroundStyle(.secondary)
                }
                LabeledContent("Version") {
                    Text("v\(item.version)")
                        .foregroundStyle(.secondary)
                }
            }
        }
        .formStyle(.grouped)
        .toolbar {
            ToolbarItemGroup {
                Button("Versions", systemImage: "clock.arrow.circlepath") {
                    loadVersionsAndShow()
                }
                .help("View version history")

                Button("Edit", systemImage: "pencil") {
                    onEdit()
                }
                .help("Edit item")

                Button("Delete", systemImage: "trash", role: .destructive) {
                    showDeleteConfirmation = true
                }
                .help("Delete item")
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

    @ViewBuilder
    private func fieldRow(key: String, value: String) -> some View {
        let isSensitive = sensitiveKeys.contains(key)
        let isRevealed = revealedFields.contains(key)
        let isCopied = copiedFieldKey == key

        LabeledContent(key.replacingOccurrences(of: "_", with: " ").capitalized) {
            HStack(spacing: 6) {
                if isSensitive && !isRevealed {
                    Text("●●●●●●●●●●●●")
                        .font(.system(.body, design: .monospaced))
                        .foregroundStyle(.tertiary)
                        .accessibilityLabel("\(key), hidden")
                } else {
                    Text(value)
                        .font(isSensitive ? .system(.body, design: .monospaced) : .body)
                        .textSelection(.enabled)
                        .accessibilityLabel("\(key), \(value)")
                }

                Spacer()

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
                    .accessibilityLabel(isRevealed ? "Hide \(key)" : "Reveal \(key)")
                }

                Button {
                    copyField(key: key, value: value)
                } label: {
                    Image(systemName: isCopied ? "checkmark" : "doc.on.doc")
                        .foregroundStyle(isCopied ? .green : .secondary)
                }
                .buttonStyle(.borderless)
                .help("Copy \(key)")
                .accessibilityLabel("Copy \(key)")
                .animation(.easeInOut(duration: 0.2), value: isCopied)
            }
        }
    }

    private func copyField(key: String, value: String) {
        let secs = vault.clipboardAutoClearEnabled ? vault.clipboardAutoClearSeconds : 0
        ClipboardService.shared.copySensitive(value, clearAfterSeconds: secs)
        copiedFieldName = key.replacingOccurrences(of: "_", with: " ")

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

private struct FlowLayout: Layout {
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
