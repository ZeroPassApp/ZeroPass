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
        VStack(alignment: .leading, spacing: 8) {
            // Header
            HStack(spacing: 6) {
                Image(systemName: vault.isUnlocked ? "lock.open.fill" : "lock.fill")
                    .foregroundStyle(vault.isUnlocked ? .green : .secondary)
                    .font(.system(size: 11))

                Text(vault.isUnlocked ? "Unlocked" : "Locked")
                    .font(.system(size: 11, weight: .semibold))

                Spacer()

                Button {
                    NSApp.activate(ignoringOtherApps: true)
                } label: {
                    Label("Open", systemImage: "macwindow")
                        .font(.system(size: 11))
                }
                .buttonStyle(.borderless)
            }

            Divider()

            if vault.isUnlocked {
                // Quick Search button - prominent
                Button {
                    quickSearch.toggle(vault: vault)
                } label: {
                    HStack(spacing: 6) {
                        Image(systemName: "magnifyingglass")
                        Text("Quick Search")
                        Spacer()
                        Text("⌘K")
                            .font(.system(size: 10))
                            .foregroundStyle(.tertiary)
                    }
                    .font(.system(size: 12))
                    .padding(.vertical, 4)
                    .padding(.horizontal, 8)
                    .background(.quaternary)
                    .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
                }
                .buttonStyle(.plain)

                // Search field
                TextField("Filter items…", text: $query)
                    .textFieldStyle(.roundedBorder)
                    .font(.system(size: 12))

                // Items list
                if items.isEmpty {
                    Text("No items found")
                        .font(.system(size: 11))
                        .foregroundStyle(.secondary)
                        .padding(.vertical, 4)
                } else {
                    ForEach(items) { item in
                        MenuBarItemRow(item: item, onCopyUsername: {
                            copyUsername(item)
                        }, onCopySecret: {
                            copySecret(item)
                        })
                    }
                }

                Divider()

                // Footer actions
                HStack {
                    Spacer()
                    Button {
                        Task { await vault.lock() }
                    } label: {
                        Label("Lock", systemImage: "lock")
                            .font(.system(size: 11))
                    }
                    .buttonStyle(.borderless)
                }
            } else {
                Text("Unlock to view items")
                    .font(.system(size: 12))
                    .foregroundStyle(.secondary)
                    .padding(.vertical, 8)

                HStack {
                    Button("Unlock") {
                        NSApp.activate(ignoringOtherApps: true)
                    }
                    Spacer()
                    Button("Quit") {
                        NSApp.terminate(nil)
                    }
                }
                .font(.system(size: 12))
            }
        }
        .padding(12)
        .frame(width: 300)
    }

    private func copyUsername(_ item: VaultItem) {
        let candidates = ["username", "email", "user", "login", "cardholder", "full_name"]
        guard let value = candidates.compactMap({ item.fields[$0] }).first(where: { !$0.isEmpty }) else { return }

        let secs = (vault.clipboardAutoClearEnabled ? vault.clipboardAutoClearSeconds : 0)
        ClipboardService.shared.copySensitive(value, clearAfterSeconds: secs)
    }

    private func copySecret(_ item: VaultItem) {
        let candidates = ["password", "api_secret", "secret", "api_key", "private_key", "cvv"]
        guard let value = candidates.compactMap({ item.fields[$0] }).first(where: { !$0.isEmpty }) else { return }

        let secs = (vault.clipboardAutoClearEnabled ? vault.clipboardAutoClearSeconds : 0)
        ClipboardService.shared.copySensitive(value, clearAfterSeconds: secs)
    }
}

private struct MenuBarItemRow: View {
    let item: VaultItem
    let onCopyUsername: () -> Void
    let onCopySecret: () -> Void

    var body: some View {
        HStack(spacing: 8) {
            Image(systemName: item.type.symbolName)
                .foregroundStyle(item.type.color)
                .font(.system(size: 11))
                .frame(width: 16)

            Text(item.name.isEmpty ? "(Untitled)" : item.name)
                .font(.system(size: 12))
                .lineLimit(1)

            Spacer()

            Button {
                onCopyUsername()
            } label: {
                Image(systemName: "person")
                    .font(.system(size: 10))
            }
            .buttonStyle(.borderless)
            .help("Copy username")

            Button {
                onCopySecret()
            } label: {
                Image(systemName: "key")
                    .font(.system(size: 10))
            }
            .buttonStyle(.borderless)
            .help("Copy secret")
        }
        .padding(.vertical, 3)
        .padding(.horizontal, 4)
        .contentShape(Rectangle())
    }
}
