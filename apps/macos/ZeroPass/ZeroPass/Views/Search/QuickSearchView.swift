import SwiftUI

struct QuickSearchView: View {
    @EnvironmentObject var vault: VaultClient
    @EnvironmentObject var panel: QuickSearchPanelController
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    @State private var query: String = ""
    @State private var results: [VaultItem] = []
    @State private var selectedID: VaultItem.ID?

    @FocusState private var focused: Bool

    @State private var searchTask: Task<Void, Never>?

    var body: some View {
        VStack(spacing: 0) {
            HStack(spacing: ZPTheme.spacing10) {
                Image(systemName: "magnifyingglass")
                    .foregroundStyle(ZPTheme.textSecondary)

                TextField("Search ZeroPass", text: $query)
                    .textFieldStyle(.plain)
                    .font(.system(size: 15, weight: .medium))
                    .foregroundStyle(ZPTheme.textPrimary)
                    .focused($focused)
                    .onSubmit { copySecret() }
            }
            .padding(.horizontal, ZPTheme.spacing16)
            .padding(.vertical, ZPTheme.spacing14)
            .background(ZPTheme.inputBackground, in: RoundedRectangle(cornerRadius: ZPTheme.radiusLarge, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: ZPTheme.radiusLarge, style: .continuous)
                    .stroke(ZPTheme.panelBorderStrong, lineWidth: 1)
            )
            .padding(.horizontal, ZPTheme.spacing14)
            .padding(.top, ZPTheme.spacing14)

            if results.isEmpty {
                ContentUnavailableView("No results", systemImage: "magnifyingglass")
                    .padding(.top, ZPTheme.spacing24)
                    .padding(.bottom, ZPTheme.spacing32)
            } else {
                List(selection: $selectedID) {
                    if query.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
                        Section("RECENT") {
                            resultRows
                        }
                    } else {
                        resultRows
                    }
                }
                .listStyle(.plain)
                .scrollContentBackground(.hidden)
                .padding(.horizontal, ZPTheme.spacing8)
                .padding(.top, ZPTheme.spacing10)
            }

            Divider()
                .overlay(ZPTheme.separatorSubtle)

            HStack(spacing: ZPTheme.spacing12) {
                shortcutHint("⏎", title: "Copy")
                shortcutHint("⌘⏎", title: "Open")
                shortcutHint("⇧⏎", title: "Open URL")
                shortcutHint("Esc", title: "Close")

                Spacer()
            }
            .padding(.horizontal, ZPTheme.spacing16)
            .padding(.vertical, ZPTheme.spacing12)
        }
        .background(.ultraThickMaterial, in: RoundedRectangle(cornerRadius: 24, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: 24, style: .continuous)
                .stroke(ZPTheme.panelBorderStrong, lineWidth: 1)
        )
        .shadow(color: ZPTheme.floatingShadow, radius: 24, y: 12)
        .padding(ZPTheme.spacing16)
        .onAppear {
            focused = true
            reloadResults()
        }
        .onChange(of: query) { _, _ in
            reloadResults()
        }
        .onExitCommand { panel.close() }
        .animation(reduceMotion ? nil : .spring(response: 0.25, dampingFraction: 0.8), value: results.count)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .overlay {
            // Keyboard shortcuts
            Group {
                Button(action: openInMainWindow) { EmptyView() }
                    .keyboardShortcut(.return, modifiers: [.command])

                Button(action: openURL) { EmptyView() }
                    .keyboardShortcut(.return, modifiers: [.shift])

                Button(action: copyUsername) { EmptyView() }
                    .keyboardShortcut("c", modifiers: [.command])
            }
            .frame(width: 0, height: 0)
            .opacity(0)
        }
    }

    @ViewBuilder
    private var resultRows: some View {
        ForEach(results) { item in
            SearchResultRow(item: item, isSelected: selectedID == item.id)
                .tag(item.id)
                .listRowSeparator(.hidden)
                .listRowInsets(EdgeInsets(top: 4, leading: 0, bottom: 4, trailing: 0))
                .listRowBackground(Color.clear)
                .contentShape(Rectangle())
                .accessibilityLabel("\(item.name), \(item.type.displayName)")
                .accessibilityHint("Activate to open")
                .onTapGesture {
                    selectedID = item.id
                    openInMainWindow()
                }
        }
    }

    private func reloadResults() {
        searchTask?.cancel()

        let trimmed = query.trimmingCharacters(in: .whitespacesAndNewlines)
        if trimmed.isEmpty {
            results = vault.recentItems(limit: 5)
            selectedID = results.first?.id
            return
        }

        searchTask = Task {
            try? await Task.sleep(nanoseconds: 150_000_000)
            guard !Task.isCancelled else { return }
            do {
                let r = try await vault.searchItems(query: trimmed)
                results = r
                selectedID = r.first?.id
            } catch {
                results = []
            }
        }
    }

    private var selectedItem: VaultItem? {
        let id = selectedID ?? results.first?.id
        guard let id else { return nil }
        return results.first(where: { $0.id == id })
    }

    private func openInMainWindow() {
        guard let id = selectedItem?.id else { return }
        vault.selectedItemID = id
        NSApp.activate(ignoringOtherApps: true)
        panel.close()
    }

    private func openURL() {
        guard let item = selectedItem else { return }
        guard let raw = item.preferredURLField?.value else { return }
        let s = raw.hasPrefix("http://") || raw.hasPrefix("https://") ? raw : "https://\(raw)"
        guard let url = URL(string: s) else { return }
        NSWorkspace.shared.open(url)
        panel.close()
    }

    private func copyUsername() {
        guard let item = selectedItem else { return }
        guard let username = item.preferredIdentityField?.value else { return }
        let secs = (vault.clipboardAutoClearEnabled ? vault.clipboardAutoClearSeconds : 0)
        ClipboardService.shared.copySensitive(username, clearAfterSeconds: secs)
        panel.close()
    }

    private func copySecret() {
        guard let item = selectedItem else { return }
        guard let value = item.preferredSecretField?.value else {
            openInMainWindow()
            return
        }

        let secs = (vault.clipboardAutoClearEnabled ? vault.clipboardAutoClearSeconds : 0)
        ClipboardService.shared.copySensitive(value, clearAfterSeconds: secs)
        panel.close()
    }

    @ViewBuilder
    private func shortcutHint(_ key: String, title: String) -> some View {
        HStack(spacing: ZPTheme.spacing6) {
            Text(key)
                .font(.system(size: 10, weight: .bold, design: .rounded))
                .foregroundStyle(ZPTheme.textPrimary)
                .padding(.horizontal, ZPTheme.spacing6)
                .padding(.vertical, ZPTheme.spacing4)
                .background(ZPTheme.chipBackground, in: RoundedRectangle(cornerRadius: 8, style: .continuous))

            Text(title)
                .font(.system(size: 10, weight: .medium))
                .foregroundStyle(ZPTheme.textSecondary)
        }
    }
}
