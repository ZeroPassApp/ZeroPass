import SwiftUI

struct SearchResultRow: View {
    let item: VaultItem
    let isSelected: Bool

    var body: some View {
        HStack(spacing: ZPTheme.spacing12) {
            ZStack {
                RoundedRectangle(cornerRadius: ZPTheme.radiusMedium, style: .continuous)
                    .fill(item.type.color.opacity(isSelected ? 0.22 : 0.12))

                Image(systemName: item.type.symbolName)
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundStyle(item.type.color)
            }
            .frame(width: 30, height: 30)

            VStack(alignment: .leading, spacing: ZPTheme.spacing4) {
                Text(item.name.isEmpty ? "(Untitled)" : item.name)
                    .font(.system(size: 13, weight: .semibold))
                    .foregroundStyle(ZPTheme.textPrimary)

                Text(item.subtitle)
                    .font(.system(size: 11, weight: .medium))
                    .foregroundStyle(ZPTheme.textSecondary)
                    .lineLimit(1)
            }

            Spacer()

            VStack(alignment: .trailing, spacing: ZPTheme.spacing4) {
                Text(item.type.displayName.uppercased())
                    .font(.system(size: 9, weight: .bold))
                    .foregroundStyle(isSelected ? ZPTheme.accent : ZPTheme.textSecondary)
                    .padding(.horizontal, ZPTheme.spacing8)
                    .padding(.vertical, ZPTheme.spacing4)
                    .background(ZPTheme.pillBackground, in: Capsule())
                    .overlay(
                        Capsule()
                            .stroke(ZPTheme.pillBorder.opacity(isSelected ? 1 : 0.55), lineWidth: 1)
                    )

                if item.favorite {
                    Image(systemName: "star.fill")
                        .font(.system(size: 10, weight: .semibold))
                        .foregroundStyle(.yellow)
                }
            }
        }
        .padding(.vertical, ZPTheme.spacing10)
        .padding(.horizontal, ZPTheme.spacing12)
        .background(
            RoundedRectangle(cornerRadius: ZPTheme.radiusLarge, style: .continuous)
                .fill(isSelected ? ZPTheme.selectionFill : Color.clear)
        )
        .overlay(
            RoundedRectangle(cornerRadius: ZPTheme.radiusLarge, style: .continuous)
                .stroke(isSelected ? ZPTheme.selectionStroke : Color.clear, lineWidth: 1)
        )
    }
}

private extension VaultItem {
    var subtitle: String {
        if let u = fields["username"], !u.isEmpty { return u }
        if let url = fields["url"], !url.isEmpty { return url }
        return notes
    }
}
