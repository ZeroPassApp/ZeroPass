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
            Section("Library") {
                sidebarRow(
                    title: "All Items",
                    systemImage: "tray.full",
                    count: vault.items.count,
                    tag: .all
                )

                sidebarRow(
                    title: "Favorites",
                    systemImage: "star.fill",
                    count: vault.favoriteCount,
                    tag: .favorites
                )
            }

            Section("Categories") {
                ForEach(VaultItemType.allCases) { type in
                    categoryRow(for: type)
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

    @ViewBuilder
    private func sidebarRow(
        title: String,
        systemImage: String,
        count: Int,
        tag: SidebarCategory
    ) -> some View {
        if count > 0 {
            Label(title, systemImage: systemImage)
                .badge(count)
                .tag(tag)
                .accessibilityLabel("\(title), \(count) items")
        } else {
            Label(title, systemImage: systemImage)
                .tag(tag)
                .foregroundStyle(.secondary)
                .accessibilityLabel("\(title), 0 items")
        }
    }

    @ViewBuilder
    private func categoryRow(for type: VaultItemType) -> some View {
        let count = vault.count(for: type)

        if count > 0 {
            Label(type.displayName, systemImage: type.symbolName)
                .badge(count)
                .tag(SidebarCategory.type(type))
                .accessibilityLabel("\(type.displayName), \(count) items")
        } else {
            Label(type.displayName, systemImage: type.symbolName)
                .tag(SidebarCategory.type(type))
                .foregroundStyle(.secondary)
                .accessibilityLabel("\(type.displayName), 0 items")
        }
    }
}
