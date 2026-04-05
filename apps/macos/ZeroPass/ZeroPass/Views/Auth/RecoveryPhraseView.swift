import SwiftUI

struct RecoveryPhraseView: View {
    @EnvironmentObject var vault: VaultClient
    let mnemonic: String

    @State private var copied = false
    @State private var confirmedSaved = false

    private var recoveryCopyGuidance: String {
        if vault.clipboardAutoClearEnabled {
            return "A paper backup is safest. If you copy this phrase digitally, the clipboard clears in \(vault.clipboardAutoClearSeconds) seconds. Ensure the destination is encrypted and offline."
        }

        return "A paper backup is safest. If you copy this phrase digitally, the clipboard will not auto-clear automatically. Ensure the destination is encrypted and offline."
    }

    var body: some View {
        AuthSceneScaffold(
            title: "Save Recovery Phrase",
            subtitle: "Write these words down before you continue to your vault.",
            detail: "You’ll need this phrase if you ever lose your master password.",
            symbolName: "key.horizontal.fill",
            symbolTint: .orange
        ) {
            VStack(alignment: .leading, spacing: ZPTheme.spacing14) {
                AuthMessageView(
                    text: "This phrase is shown only now. Store it somewhere secure and offline if possible.",
                    systemImage: "exclamationmark.shield.fill",
                    tone: .warning
                )

                RecoveryPhraseCardView(mnemonic: mnemonic)

                Text(recoveryCopyGuidance)
                    .font(.footnote)
                    .foregroundStyle(ZPTheme.textSecondary)
                    .fixedSize(horizontal: false, vertical: true)

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
}
