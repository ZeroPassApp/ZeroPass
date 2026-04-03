import SwiftUI

struct MainShellView: View {
    @EnvironmentObject var vault: VaultClient

    @State private var selectedCategory: SidebarCategory? = .all
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
                .navigationSplitViewColumnWidth(min: 180, ideal: 200)
        } content: {
            ItemListView(
                category: selectedCategory,
                selectedItemID: $vault.selectedItemID,
                editingItem: $editingItem
            )
            .environmentObject(vault)
            .navigationSplitViewColumnWidth(min: 220, ideal: 280)
        } detail: {
            if let item = selectedItem {
                ItemDetailView(item: item, onEdit: { editingItem = item })
                    .environmentObject(vault)
                    .id(item.id)
            } else {
                ContentUnavailableView("Select an Item", systemImage: "tray", description: Text("Choose an item from the list to view its details."))
            }
        }
        .navigationTitle(vault.vaultName)
        .toolbar {
            ToolbarItemGroup {
                Button {
                    Task { try? await vault.refreshItems() }
                } label: {
                    Image(systemName: "arrow.clockwise")
                        .accessibilityLabel("Refresh items")
                }
                .help("Refresh items")
                .keyboardShortcut("r", modifiers: [.command])

                Button {
                    showingNewItem = true
                } label: {
                    Image(systemName: "plus")
                        .accessibilityLabel("New item")
                }
                .help("New item")
                .keyboardShortcut("n", modifiers: [.command])

                Button {
                    Task { await vault.lock() }
                } label: {
                    Image(systemName: "lock")
                        .accessibilityLabel("Lock vault")
                }
                .help("Lock vault")
            }
        }
        .sheet(isPresented: $showingNewItem) {
            ItemEditorView(item: VaultItem.new(type: .login))
                .environmentObject(vault)
        }
        .sheet(item: $editingItem) { item in
            ItemEditorView(item: item)
                .environmentObject(vault)
        }
    }
}
