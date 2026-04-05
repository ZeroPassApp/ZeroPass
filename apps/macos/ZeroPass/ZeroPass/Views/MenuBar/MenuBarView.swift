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
            HStack(spacing: ZPTheme.spacing8) {
                ZStack {
                    Circle()
                        .fill((vault.isUnlocked ? ZPTheme.success : ZPTheme.textMuted).opacity(0.18))

                    Image(systemName: vault.isUnlocked ? "lock.open.fill" : "lock.fill")
                        .foregroundStyle(vault.isUnlocked ? ZPTheme.success : ZPTheme.textSecondary)
                        .font(.system(size: 11, weight: .semibold))
                }
                .frame(width: 24, height: 24)

                VStack(alignment: .leading, spacing: 2) {
                    Text(vault.vaultName)
                        .font(.system(size: 12, weight: .semibold))
                        .foregroundStyle(ZPTheme.textPrimary)

                    Text(vault.isUnlocked ? "Unlocked" : "Locked")
                        .font(.system(size: 10, weight: .medium))
                        .foregroundStyle(ZPTheme.textSecondary)
                }

                Spacer()

                Button {
                    NSApp.activate(ignoringOtherApps: true)
                } label: {
                    Image(systemName: "macwindow")
                        .font(.system(size: 11, weight: .semibold))
                }
                .buttonStyle(.borderless)
                .accessibilityLabel("Open main window")
            }
            .padding(.horizontal, ZPTheme.spacing12)
            .padding(.vertical, ZPTheme.spacing10)
            .zpSurface(.muted, radius: ZPTheme.radiusLarge, shadow: false)

            Divider()

            if vault.isUnlocked {
                Button {
                    quickSearch.toggle(vault: vault)
                } label: {
                    HStack(spacing: ZPTheme.spacing8) {
                        Image(systemName: "magnifyingglass")
                        Text("Quick Search")
                        Spacer()
                        Text("⌘K")
                            .font(.system(size: 10))
                            .foregroundStyle(ZPTheme.textTertiary)
                    }
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundStyle(ZPTheme.textPrimary)
                    .padding(.vertical, ZPTheme.spacing10)
                    .padding(.horizontal, ZPTheme.spacing12)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .zpSurface(.accent, radius: ZPTheme.radiusLarge, shadow: false)
                }
                .buttonStyle(.plain)
                .accessibilityLabel("Quick Search, Command K")

                TextField("Filter items…", text: $query)
                    .textFieldStyle(.plain)
                    .font(.system(size: 12, weight: .medium))
                    .padding(.horizontal, ZPTheme.spacing12)
                    .padding(.vertical, ZPTheme.spacing10)
                    .zpSurface(.inset, radius: ZPTheme.radiusLarge, shadow: false)

                if items.isEmpty {
                    Text("No items found")
                        .font(.system(size: 11))
                        .foregroundStyle(ZPTheme.textSecondary)
                        .padding(.vertical, ZPTheme.spacing8)
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

                HStack {
                    Spacer()
                    Button {
                        Task { await vault.lock() }
                    } label: {
                        Label("Lock", systemImage: "lock")
                            .font(.system(size: 11, weight: .semibold))
                    }
                    .buttonStyle(.borderless)
                    .accessibilityLabel("Lock vault")
                }
            } else {
                Text("Unlock to view items")
                    .font(.system(size: 12))
                    .foregroundStyle(ZPTheme.textSecondary)
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
        .frame(width: 320)
        .background(ZPTheme.workspaceBackground)
    }

    private func copyUsername(_ item: VaultItem) {
        guard let value = item.preferredIdentityField?.value else { return }

        let secs = (vault.clipboardAutoClearEnabled ? vault.clipboardAutoClearSeconds : 0)
        ClipboardService.shared.copySensitive(value, clearAfterSeconds: secs)
    }

    private func copySecret(_ item: VaultItem) {
        guard let value = item.preferredSecretField?.value else { return }

        let secs = (vault.clipboardAutoClearEnabled ? vault.clipboardAutoClearSeconds : 0)
        ClipboardService.shared.copySensitive(value, clearAfterSeconds: secs)
    }
}

private struct MenuBarItemRow: View {
    let item: VaultItem
    let onCopyUsername: () -> Void
    let onCopySecret: () -> Void

    var body: some View {
        HStack(spacing: ZPTheme.spacing10) {
            ZStack {
                RoundedRectangle(cornerRadius: 8, style: .continuous)
                    .fill(item.type.color.opacity(0.12))

                Image(systemName: item.type.symbolName)
                    .foregroundStyle(item.type.color)
                    .font(.system(size: 11, weight: .semibold))
            }
            .frame(width: 24, height: 24)

            Text(item.name.isEmpty ? "(Untitled)" : item.name)
                .font(.system(size: 12, weight: .medium))
                .foregroundStyle(ZPTheme.textPrimary)
                .lineLimit(1)

            Spacer()

            Button {
                onCopyUsername()
            } label: {
                Image(systemName: "person")
                    .font(.system(size: 10))
            }
            .buttonStyle(.borderless)
            .accessibilityLabel("Copy username")
            .help("Copy username")

            Button {
                onCopySecret()
            } label: {
                Image(systemName: "key")
                    .font(.system(size: 10))
            }
            .buttonStyle(.borderless)
            .accessibilityLabel("Copy secret")
            .help("Copy secret")
        }
        .padding(.vertical, ZPTheme.spacing8)
        .padding(.horizontal, ZPTheme.spacing10)
        .zpSurface(.card, radius: ZPTheme.radiusLarge, shadow: false)
        .contentShape(Rectangle())
    }
}
