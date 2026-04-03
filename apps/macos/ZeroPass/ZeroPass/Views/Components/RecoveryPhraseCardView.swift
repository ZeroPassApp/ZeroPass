import SwiftUI

struct RecoveryPhraseCardView: View {
    let mnemonic: String

    private var words: [String] {
        RecoveryPhraseSupport.normalizedWords(from: mnemonic)
    }

    private var columns: [GridItem] {
        let count = words.count > 12 ? 4 : 3
        return Array(repeating: GridItem(.flexible(), spacing: ZPTheme.spacing8), count: max(count, 1))
    }

    var body: some View {
        if words.isEmpty {
            Text("No recovery phrase available.")
                .font(.callout)
                .foregroundStyle(ZPTheme.textSecondary)
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(ZPTheme.spacing16)
                .background(ZPTheme.authInsetBackground, in: RoundedRectangle(cornerRadius: ZPTheme.radiusLarge, style: .continuous))
                .overlay(
                    RoundedRectangle(cornerRadius: ZPTheme.radiusLarge, style: .continuous)
                        .stroke(ZPTheme.authInsetBorder, lineWidth: 1)
                )
        } else {
            LazyVGrid(columns: columns, spacing: ZPTheme.spacing8) {
                ForEach(Array(words.enumerated()), id: \.offset) { index, word in
                    HStack(alignment: .firstTextBaseline, spacing: ZPTheme.spacing6) {
                        Text("\(index + 1).")
                            .font(.caption)
                            .foregroundStyle(ZPTheme.textSecondary)
                            .frame(width: 22, alignment: .trailing)

                        Text(word)
                            .font(.system(.body, design: .rounded))
                            .fontWeight(.medium)
                            .foregroundStyle(ZPTheme.textPrimary)
                    }
                    .padding(.vertical, ZPTheme.spacing8)
                    .padding(.horizontal, ZPTheme.spacing10)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .background(ZPTheme.authInsetBackground, in: RoundedRectangle(cornerRadius: ZPTheme.radiusMedium, style: .continuous))
                    .overlay(
                        RoundedRectangle(cornerRadius: ZPTheme.radiusMedium, style: .continuous)
                            .stroke(ZPTheme.authInsetBorder, lineWidth: 1)
                    )
                    .accessibilityLabel("Word \(index + 1), \(word)")
                }
            }
            .textSelection(.enabled)
        }
    }
}
