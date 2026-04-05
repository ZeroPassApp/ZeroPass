import SwiftUI

struct RecoveryPhraseView: View {
    @EnvironmentObject var vault: VaultClient
    let mnemonic: String

    @State private var copied = false
    @State private var confirmedSaved = false

    var body: some View {
        AuthSceneScaffold(
            title: "Save Recovery Phrase",
            subtitle: "Write these words down before you continue to your vault.",
            detail: "You’ll need this phrase if you ever lose your master password.",
            symbolName: "key.horizontal.fill",
            symbolTint: .orange
        ) {
            VStack(alignment: .leading, spacing: ZPTheme.spacing16) {
                AuthMessageView(
                    text: "This phrase is shown only now. Store it somewhere secure and offline if possible.",
                    systemImage: "exclamationmark.shield.fill",
                    tone: .warning
                )

                RecoveryPhraseCardView(mnemonic: mnemonic)

                VStack(alignment: .leading, spacing: ZPTheme.spacing12) {
                    Text("A paper backup is safest. If you copy this phrase digitally, ensure the destination is encrypted and offline.")
                        .font(.footnote)
                        .foregroundStyle(ZPTheme.textSecondary)
                        .fixedSize(horizontal: false, vertical: true)

                    HStack(spacing: ZPTheme.spacing8) {
                        recoveryPill("One-Time Display", systemImage: "eye.slash")
                        recoveryPill(vault.clipboardAutoClearEnabled ? "Clipboard clears in \(vault.clipboardAutoClearSeconds)s" : "Clipboard not auto-cleared", systemImage: "doc.on.clipboard")
                    }
                }

                HStack(spacing: ZPTheme.spacing12) {
                    Button {
                        let secs = vault.clipboardAutoClearEnabled ? vault.clipboardAutoClearSeconds : 0
                        ClipboardService.shared.copySensitive(mnemonic, clearAfterSeconds: secs)
                        copied = true
                        DispatchQueue.main.asyncAfter(deadline: .now() + 2) { copied = false }
                    } label: {
                        Label(copied ? "Copied" : "Copy", systemImage: copied ? "checkmark" : "doc.on.doc")
                    }
                    .buttonStyle(.bordered)
                    .accessibilityIdentifier("recoveryPhrase.copyButton")

                    Spacer()

                    Button {
                        vault.acceptRecoveryPhrase()
                    } label: {
                        Text("Continue")
                    }
                    .controlSize(.large)
                    .buttonStyle(.borderedProminent)
                    .disabled(!confirmedSaved)
                    .keyboardShortcut(.defaultAction)
                    .accessibilityHint("Continue after you have saved the recovery phrase.")
                    .accessibilityIdentifier("recoveryPhrase.continueButton")
                }

                Toggle("I have saved this recovery phrase", isOn: $confirmedSaved)
                    .accessibilityIdentifier("recoveryPhrase.confirmSavedToggle")
            }
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
