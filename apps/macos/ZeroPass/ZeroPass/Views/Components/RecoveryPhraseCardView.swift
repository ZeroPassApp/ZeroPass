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
                .zpSurface(.muted)
        } else {
            VStack(alignment: .leading, spacing: ZPTheme.spacing14) {
                HStack(alignment: .top) {
                    VStack(alignment: .leading, spacing: ZPTheme.spacing4) {
                        Text("Recovery Phrase")
                            .font(.headline)
                            .foregroundStyle(ZPTheme.textPrimary)

                        Text("Store these words offline. ZeroPass cannot recover them for you.")
                            .font(.caption)
                            .foregroundStyle(ZPTheme.textSecondary)
                    }

                    Spacer()

                    Text("\(words.count) words")
                        .font(.caption.weight(.semibold))
                        .foregroundStyle(ZPTheme.accent)
                        .padding(.horizontal, ZPTheme.spacing10)
                        .padding(.vertical, ZPTheme.spacing6)
                        .background(ZPTheme.pillBackground, in: Capsule())
                        .overlay(
                            Capsule()
                                .stroke(ZPTheme.pillBorder, lineWidth: 1)
                        )
                }

                LazyVGrid(columns: columns, spacing: ZPTheme.spacing8) {
                    ForEach(Array(words.enumerated()), id: \.offset) { index, word in
                        HStack(alignment: .firstTextBaseline, spacing: ZPTheme.spacing8) {
                            Text("\(index + 1)")
                                .font(.caption.weight(.bold))
                                .foregroundStyle(ZPTheme.accent)
                                .frame(width: 24, alignment: .trailing)

                            Text(word)
                                .font(.system(.body, design: .rounded))
                                .fontWeight(.medium)
                                .foregroundStyle(ZPTheme.textPrimary)
                        }
                        .padding(.vertical, ZPTheme.spacing8)
                        .padding(.horizontal, ZPTheme.spacing12)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .zpSurface(.inset, radius: ZPTheme.radiusMedium, shadow: false)
                        .accessibilityLabel("Word \(index + 1), \(word)")
                    }
                }

                HStack(spacing: ZPTheme.spacing8) {
                    recoveryPill("Shown Once", systemImage: "eye.slash")
                    recoveryPill("Save Offline", systemImage: "tray.and.arrow.down")
                }
            }
            .textSelection(.enabled)
            .privacySensitive()
            .padding(ZPTheme.spacing16)
            .zpSurface(.elevated)
        }
    }

    @ViewBuilder
    private func recoveryPill(_ title: String, systemImage: String) -> some View {
        Label(title, systemImage: systemImage)
            .font(.caption.weight(.semibold))
            .foregroundStyle(ZPTheme.textSecondary)
            .padding(.horizontal, ZPTheme.spacing10)
            .padding(.vertical, ZPTheme.spacing8)
            .background(ZPTheme.chipBackground, in: Capsule())
            .overlay(
                Capsule()
                    .stroke(ZPTheme.panelBorder, lineWidth: 1)
            )
    }
}
