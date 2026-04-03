import SwiftUI

enum SidebarCategory: Hashable {
    case all
    case favorites
    case type(VaultItemType)
    case tag(String)
}

struct SidebarView: View {
    @EnvironmentObject var vault: VaultClient
    @Binding var selectedCategory: SidebarCategory?

    var body: some View {
        List(selection: $selectedCategory) {
            Section {
                Label("All Items", systemImage: "tray.full")
                    .badge(vault.items.count)
                    .tag(SidebarCategory.all)
                    .accessibilityLabel("All Items, \(vault.items.count) items")

                Label("Favorites", systemImage: "star.fill")
                    .foregroundStyle(.yellow, .primary)
                    .badge(vault.favoriteCount)
                    .tag(SidebarCategory.favorites)
                    .accessibilityLabel("Favorites, \(vault.favoriteCount) items")
            }

            Section("Categories") {
                ForEach(VaultItemType.allCases) { type in
                    Label(type.displayName, systemImage: type.symbolName)
                        .foregroundStyle(type.color, .primary)
                        .badge(vault.count(for: type))
                        .tag(SidebarCategory.type(type))
                        .accessibilityLabel("\(type.displayName), \(vault.count(for: type)) items")
                }
            }

            if !vault.allTags.isEmpty {
                Section("Tags") {
                    ForEach(vault.allTags, id: \.self) { tag in
                        Label(tag, systemImage: "tag")
                            .tag(SidebarCategory.tag(tag))
                            .accessibilityLabel("Tag: \(tag)")
                    }
                }
            }
        }
        .listStyle(.sidebar)
        .accessibilityElement(children: .contain)
        .accessibilityLabel("Vault categories")
    }
}
