import SwiftUI

struct ItemDetailView: View {
    @EnvironmentObject var vault: VaultClient

    let item: VaultItem
    let onEdit: () -> Void

    @State private var showingVersions = false
    @State private var versions: [ItemVersion] = []
    @State private var versionsError: String?

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack {
                Text(item.name.isEmpty ? "(Untitled)" : item.name)
                    .font(.title2)
                    .bold()

                Spacer()

                Button("Edit") { onEdit() }
            }

            Text(item.type.displayName)
                .foregroundStyle(.secondary)

            if !item.notes.isEmpty {
                Text(item.notes)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .textSelection(.enabled)
            }

            if !item.fields.isEmpty {
                Form {
                    ForEach(item.fields.keys.sorted(), id: \.self) { key in
                        HStack {
                            Text(key)
                            Spacer()
                            Text(item.fields[key] ?? "")
                                .textSelection(.enabled)
                        }
                    }
                }
            }

            Spacer()

            HStack {
                Button("Versions") { loadVersionsAndShow() }

                Spacer()

                Button(role: .destructive) {
                    Task { try? await vault.deleteItem(id: item.id) }
                } label: {
                    Text("Delete")
                }
            }
        }
        .padding(20)
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
                        Text("v\(v.version)")
                        Spacer()
                        Text(v.savedAt.formatted())
                            .foregroundStyle(.secondary)
                        Button("Restore") { onRestore(v.version) }
                    }
                }
            }

            HStack {
                Spacer()
                Button("Close") { dismiss() }
            }
        }
        .padding(20)
        .frame(width: 520, height: 420)
    }
}
