import SwiftUI

struct MainShellView: View {
    @EnvironmentObject var vault: VaultClient

    @State private var selectedCategory: SidebarCategory? = .all
    @State private var searchText = ""
    @State private var showingNewItem = false
    @State private var editingItem: VaultItem?

    private var selectedItem: VaultItem? {
        guard let id = vault.selectedItemID else { return nil }
        return vault.items.first(where: { $0.id == id })
    }

    var body: some View {
        NavigationSplitView {
            SidebarView(selectedCategory: $selectedCategory)
                .environmentObject(vault)
                .navigationSplitViewColumnWidth(min: 180, ideal: 220)
                .accessibilityElement(children: .contain)
                .accessibilityLabel("Sidebar")
                .background(ZPTheme.sidebarBackground)
        } content: {
            ItemListView(
                category: selectedCategory,
                searchText: $searchText,
                selectedItemID: $vault.selectedItemID,
                editingItem: $editingItem,
                onCreateItem: { showingNewItem = true }
            )
            .environmentObject(vault)
            .navigationSplitViewColumnWidth(min: 260, ideal: 320)
            .accessibilityElement(children: .contain)
            .accessibilityLabel("Item list")
            .background(ZPTheme.panelBackgroundMuted)
        } detail: {
            if let item = selectedItem {
                ItemDetailView(item: item, onEdit: { editingItem = item })
                    .environmentObject(vault)
                    .id(item.id)
            } else if vault.items.isEmpty {
                ContentUnavailableView {
                    Label("Vault is Empty", systemImage: "tray")
                } description: {
                    Text("Create your first item to start storing secrets securely.")
                } actions: {
                    Button("New Item") {
                        showingNewItem = true
                    }
                }
                .background(ZPTheme.workspaceBackground)
            } else {
                ContentUnavailableView(
                    "Select an Item",
                    systemImage: "tray",
                    description: Text("Choose an item from the list to view its details."))
                .background(ZPTheme.workspaceBackground)
            }
        }
        .navigationTitle(vault.vaultName.isEmpty ? "ZeroPass" : vault.vaultName)
        .searchable(
            text: $searchText,
            placement: .toolbar,
            prompt: "Names, usernames, notes, domains"
        )
        .onAppear {
            revealSelectedItemInVisibleScope()
        }
        .onChange(of: vault.selectedItemID) { _, _ in
            revealSelectedItemInVisibleScope()
        }
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button {
                    showingNewItem = true
                } label: {
                    Label("New Item", systemImage: "plus")
                }
                .help("New item")
                .keyboardShortcut("n", modifiers: [.command])
            }
        }
        .background(ZPTheme.workspaceBackground)
        .sheet(isPresented: $showingNewItem) {
            ItemEditorView(item: VaultItem.new(type: .login))
                .environmentObject(vault)
        }
        .sheet(item: $editingItem) { item in
            ItemEditorView(item: item)
                .environmentObject(vault)
        }
    }

    private func revealSelectedItemInVisibleScope() {
        guard let item = selectedItem else { return }

        if !selectedCategoryContains(item) {
            selectedCategory = .all
        }

        if !searchText.isEmpty && !VaultItemSearchMatcher.matches(item: item, query: searchText) {
            searchText = ""
        }
    }

    private func selectedCategoryContains(_ item: VaultItem) -> Bool {
        switch selectedCategory {
        case .all, .none:
            true
        case .favorites:
            item.favorite
        case .type(let type):
            item.type == type
        case .tag(let tag):
            item.tags.contains(tag)
        }
    }
}
