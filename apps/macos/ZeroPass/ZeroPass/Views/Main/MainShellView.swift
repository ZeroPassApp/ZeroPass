import SwiftUI

struct MainShellView: View {
    @EnvironmentObject var vault: VaultClient

    @State private var showingNewItem = false
    @State private var editingItem: VaultItem?

    private var selectedItem: VaultItem? {
        guard let id = vault.selectedItemID else { return nil }
        return vault.items.first(where: { $0.id == id })
    }

    var body: some View {
        NavigationSplitView {
            List(selection: $vault.selectedItemID) {
                Section("Items") {
                    ForEach(vault.items) { item in
                        HStack(spacing: 8) {
                            if item.favorite {
                                Image(systemName: "star.fill")
                                    .foregroundStyle(.yellow)
                            }
                            Text(item.name.isEmpty ? "(Untitled)" : item.name)
                        }
                        .tag(item.id)
                        .contextMenu {
                            Button("Edit") { editingItem = item }
                            Button(role: .destructive) {
                                Task { try? await vault.deleteItem(id: item.id) }
                            } label: {
                                Text("Delete")
                            }
                        }
                    }
                }
            }
            .navigationTitle("ZeroPass")
        } detail: {
            if let item = selectedItem {
                ItemDetailView(item: item, onEdit: { editingItem = item })
            } else {
                ContentUnavailableView("No item selected", systemImage: "tray")
            }
        }
        .toolbar {
            ToolbarItemGroup {
                Button {
                    Task { try? await vault.refreshItems() }
                } label: {
                    Image(systemName: "arrow.clockwise")
                }

                Button {
                    showingNewItem = true
                } label: {
                    Image(systemName: "plus")
                }

                Button {
                    Task { await vault.lock() }
                } label: {
                    Image(systemName: "lock")
                }
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
