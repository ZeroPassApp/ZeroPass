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
        VStack(spacing: 12) {
            HStack(spacing: 10) {
                Image(systemName: "magnifyingglass")
                    .foregroundStyle(.secondary)

                TextField("Search", text: $query)
                    .textFieldStyle(.plain)
                    .font(.system(size: 15))
                    .focused($focused)
                    .onSubmit { copySecret() }
            }
            .padding(.horizontal, 14)
            .padding(.vertical, 12)

            Divider()

            if results.isEmpty {
                ContentUnavailableView("No results", systemImage: "magnifyingglass")
                    .padding(.top, 18)
            } else {
                List(selection: $selectedID) {
                    if query.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
                        Section("RECENT") {
                            ForEach(results) { item in
                                SearchResultRow(item: item, isSelected: selectedID == item.id)
                                    .tag(item.id)
                                    .listRowSeparator(.hidden)
                                    .listRowInsets(EdgeInsets())
                                    .contentShape(Rectangle())
                                    .accessibilityLabel("\(item.name), \(item.type.displayName)")
                                    .accessibilityHint("Activate to open")
                                    .onTapGesture {
                                        selectedID = item.id
                                        openInMainWindow()
                                    }
                            }
                        }
                    } else {
                        ForEach(results) { item in
                            SearchResultRow(item: item, isSelected: selectedID == item.id)
                                .tag(item.id)
                                .listRowSeparator(.hidden)
                                .listRowInsets(EdgeInsets())
                                .contentShape(Rectangle())
                                .accessibilityLabel("\(item.name), \(item.type.displayName)")
                                .accessibilityHint("Activate to open")
                                .onTapGesture {
                                    selectedID = item.id
                                    openInMainWindow()
                                }
                        }
                    }
                }
                .listStyle(.plain)
                .padding(.horizontal, 6)
            }

            Divider()

            HStack {
                Text("⏎ Copy secret  ⌘⏎ Open  ⇧⏎ Open URL  ⌘C Copy username  Esc Close")
                    .font(.system(size: 10))
                    .foregroundStyle(.secondary)

                Spacer()
            }
            .padding(.horizontal, 14)
            .padding(.bottom, 10)
        }
        .background(.ultraThickMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 12, style: .continuous))
        .padding(10)
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
        guard let raw = item.fields["url"], !raw.isEmpty else { return }
        let s = raw.hasPrefix("http://") || raw.hasPrefix("https://") ? raw : "https://\(raw)"
        guard let url = URL(string: s) else { return }
        NSWorkspace.shared.open(url)
        panel.close()
    }

    private func copyUsername() {
        guard let item = selectedItem else { return }
        guard let username = item.fields["username"], !username.isEmpty else { return }
        let secs = (vault.clipboardAutoClearEnabled ? vault.clipboardAutoClearSeconds : 0)
        ClipboardService.shared.copySensitive(username, clearAfterSeconds: secs)
        panel.close()
    }

    private func copySecret() {
        guard let item = selectedItem else { return }
        let candidates = ["password", "api_secret", "secret", "private_key", "cvv"]
        guard let value = candidates.compactMap({ item.fields[$0] }).first(where: { !$0.isEmpty }) else {
            openInMainWindow()
            return
        }

        let secs = (vault.clipboardAutoClearEnabled ? vault.clipboardAutoClearSeconds : 0)
        ClipboardService.shared.copySensitive(value, clearAfterSeconds: secs)
        panel.close()
    }
}
