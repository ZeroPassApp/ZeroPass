import SwiftUI

struct UnlockRecoverySection: View {
    @Binding var mnemonic: String

    let focusRequestID: UUID
    let showErrorHighlight: Bool

    @FocusState private var isRecoveryFocused: Bool

    private var summary: RecoveryPhraseSummary {
        RecoveryPhraseSupport.summary(from: mnemonic)
    }

    var body: some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing12) {
            Text("Recovery Phrase")
                .font(.callout)
                .fontWeight(.medium)
                .foregroundStyle(showErrorHighlight ? ZPTheme.destructive : ZPTheme.textSecondary)

            Text("Paste your 12-, 15-, 18-, 21-, or 24-word recovery phrase. Numbering and line breaks are okay.")
                .font(.footnote)
                .foregroundStyle(ZPTheme.textSecondary)
                .fixedSize(horizontal: false, vertical: true)

            ZStack(alignment: .topLeading) {
                if mnemonic.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
                    Text("Paste your recovery phrase here")
                        .foregroundStyle(ZPTheme.textTertiary)
                        .padding(.top, 16)
                        .padding(.horizontal, 16)
                        .allowsHitTesting(false)
                }

                TextEditor(text: $mnemonic)
                    .font(.body)
                    .scrollContentBackground(.hidden)
                    .focused($isRecoveryFocused)
                    .autocorrectionDisabled()
                    .privacySensitive()
                    .padding(8)
                    .frame(minHeight: ZPTheme.authEditorMinHeight)
                    .accessibilityLabel("Recovery phrase")
                    .accessibilityHint("Paste your recovery phrase to unlock the vault")
            }
            .padding(2)
            .background(borderColor.opacity(showErrorHighlight ? 0.12 : 0.08), in: RoundedRectangle(cornerRadius: ZPTheme.radiusLarge, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: ZPTheme.radiusLarge, style: .continuous)
                    .stroke(borderColor, lineWidth: borderWidth)
            )

            Label(summary.helperText, systemImage: summaryIcon)
                .font(.caption)
                .foregroundStyle(summaryColor)
                .fixedSize(horizontal: false, vertical: true)
        }
        .onAppear {
            DispatchQueue.main.async {
                isRecoveryFocused = true
            }
        }
        .onChange(of: focusRequestID) { _, _ in
            DispatchQueue.main.async {
                isRecoveryFocused = true
            }
        }
    }

    private var borderColor: Color {
        if showErrorHighlight {
            return ZPTheme.destructive
        }
        return isRecoveryFocused ? ZPTheme.accent : ZPTheme.authInsetBorder
    }

    private var borderWidth: CGFloat {
        (showErrorHighlight || isRecoveryFocused) ? 1.5 : 1
    }

    private var summaryIcon: String {
        if summary.isEmpty {
            return "info.circle"
        }
        return summary.isStandardLength ? "checkmark.circle.fill" : "exclamationmark.circle"
    }

    private var summaryColor: Color {
        if summary.isEmpty {
            return ZPTheme.textSecondary
        }
        return summary.isStandardLength ? ZPTheme.success : ZPTheme.warning
    }
}
