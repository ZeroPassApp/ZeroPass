import SwiftUI

struct MenuBarView: View {
    @EnvironmentObject var vault: VaultClient
    @EnvironmentObject var quickSearch: QuickSearchPanelController

    @State private var query: String = ""

    private var items: [VaultItem] {
        let trimmed = query.trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmed.isEmpty {
            return vault.recentItems(limit: 5)
        }
        return vault.items
            .filter { $0.name.localizedCaseInsensitiveContains(trimmed) }
            .prefix(8)
            .map { $0 }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack {
                Image(systemName: vault.isUnlocked ? "lock.open.fill" : "lock.fill")
                Text(vault.isUnlocked ? "Unlocked" : "Locked")
                    .font(.system(size: 12, weight: .semibold))

                Spacer()

                Button("Open") {
                    NSApp.activate(ignoringOtherApps: true)
                }
            }

            Divider()

            if vault.isUnlocked {
                TextField("Search", text: $query)
                    .textFieldStyle(.roundedBorder)

                ForEach(items) { item in
                    HStack {
                        Text(item.name.isEmpty ? "(Untitled)" : item.name)
                            .lineLimit(1)

                        Spacer()

                        Button {
                            copySecret(item)
                        } label: {
                            Image(systemName: "doc.on.doc")
                        }
                        .buttonStyle(.borderless)
                    }
                }

                Divider()

                HStack {
                    Button("Quick Search") {
                        quickSearch.toggle(vault: vault)
                    }
                    Spacer()
                    Button("Lock") {
                        Task { await vault.lock() }
                    }
                }
            } else {
                Text("Unlock to view items")
                    .foregroundStyle(.secondary)

                HStack {
                    Button("Unlock") {
                        NSApp.activate(ignoringOtherApps: true)
                    }
                    Spacer()
                    Button("Quit") {
                        NSApp.terminate(nil)
                    }
                }
            }
        }
        .padding(12)
        .frame(width: 280)
    }

    private func copySecret(_ item: VaultItem) {
        let candidates = ["password", "api_secret", "secret", "private_key", "cvv"]
        guard let value = candidates.compactMap({ item.fields[$0] }).first(where: { !$0.isEmpty }) else { return }

        let secs = (vault.clipboardAutoClearEnabled ? vault.clipboardAutoClearSeconds : 0)
        ClipboardService.shared.copySensitive(value, clearAfterSeconds: secs)
    }
}
