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
        .scrollContentBackground(.hidden)
        .background(ZPTheme.sidebarBackground)
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
        HStack(spacing: ZPTheme.spacing10) {
            Image(systemName: systemImage)
                .font(.system(size: 13, weight: .semibold))
                .foregroundStyle(count > 0 ? ZPTheme.textSecondary : ZPTheme.textMuted)
                .frame(width: 16)

            Text(title)
                .font(.system(size: 13, weight: .medium))

            Spacer()

            if count > 0 {
                Text("\(count)")
                    .font(.caption.weight(.bold))
                    .foregroundStyle(ZPTheme.textSecondary)
                    .padding(.horizontal, ZPTheme.spacing8)
                    .padding(.vertical, ZPTheme.spacing4)
                    .background(ZPTheme.chipBackground, in: Capsule())
            }
        }
        .tag(tag)
        .accessibilityLabel("\(title), \(count) items")
    }

    @ViewBuilder
    private func categoryRow(for type: VaultItemType) -> some View {
        let count = vault.count(for: type)

        HStack(spacing: ZPTheme.spacing10) {
            Image(systemName: type.symbolName)
                .font(.system(size: 13, weight: .semibold))
                .foregroundStyle(count > 0 ? type.color : ZPTheme.textMuted)
                .frame(width: 16)

            Text(type.displayName)
                .font(.system(size: 13, weight: .medium))

            Spacer()

            if count > 0 {
                Text("\(count)")
                    .font(.caption.weight(.bold))
                    .foregroundStyle(ZPTheme.textSecondary)
                    .padding(.horizontal, ZPTheme.spacing8)
                    .padding(.vertical, ZPTheme.spacing4)
                    .background(ZPTheme.chipBackground, in: Capsule())
            }
        }
        .tag(SidebarCategory.type(type))
        .accessibilityLabel("\(type.displayName), \(count) items")
    }
}
