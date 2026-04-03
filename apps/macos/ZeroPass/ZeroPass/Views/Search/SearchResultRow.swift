import SwiftUI

struct SearchResultRow: View {
    let item: VaultItem
    let isSelected: Bool

    var body: some View {
        HStack(spacing: 10) {
            Image(systemName: item.type.symbolName)
                .foregroundStyle(item.type.color)
                .frame(width: 16)

            VStack(alignment: .leading, spacing: 2) {
                Text(item.name.isEmpty ? "(Untitled)" : item.name)
                    .font(.system(size: 13, weight: .semibold))

                Text(item.subtitle)
                    .font(.system(size: 11))
                    .foregroundStyle(.secondary)
                    .lineLimit(1)
            }

            Spacer()

            Text(item.type.displayName.uppercased())
                .font(.system(size: 10, weight: .medium))
                .foregroundStyle(.secondary)
        }
        .padding(.vertical, 6)
        .padding(.horizontal, 10)
        .background(isSelected ? Color.accentColor.opacity(0.18) : Color.clear)
        .clipShape(RoundedRectangle(cornerRadius: 8))
    }
}

private extension VaultItem {
    var subtitle: String {
        if let u = fields["username"], !u.isEmpty { return u }
        if let url = fields["url"], !url.isEmpty { return url }
        return notes
    }
}
