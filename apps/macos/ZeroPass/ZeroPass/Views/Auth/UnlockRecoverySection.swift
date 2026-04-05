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
        VStack(alignment: .leading, spacing: ZPTheme.spacing8) {
            ZStack(alignment: .topLeading) {
                if mnemonic.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
                    Text("Paste Recovery Phrase")
                        .foregroundStyle(ZPTheme.textTertiary)
                        .padding(.top, 14)
                        .padding(.horizontal, 14)
                        .allowsHitTesting(false)
                }

                TextEditor(text: $mnemonic)
                    .font(.body)
                    .scrollContentBackground(.hidden)
                    .focused($isRecoveryFocused)
                    .autocorrectionDisabled()
                    .privacySensitive()
                    .padding(8)
                    .frame(minHeight: 104)
                    .accessibilityLabel("Recovery phrase")
                    .accessibilityHint("Paste your recovery phrase to unlock the vault")
            }
            .padding(2)
            .background(borderColor.opacity(showErrorHighlight ? 0.12 : 0.08), in: RoundedRectangle(cornerRadius: ZPTheme.radiusMedium, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: ZPTheme.radiusMedium, style: .continuous)
                    .stroke(borderColor, lineWidth: borderWidth)
            )

            Label(summary.helperText, systemImage: summaryIcon)
                .font(.caption2)
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
