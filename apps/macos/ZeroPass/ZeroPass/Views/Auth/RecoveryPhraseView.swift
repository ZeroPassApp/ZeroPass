import SwiftUI

struct RecoveryPhraseView: View {
    @EnvironmentObject var vault: VaultClient
    let mnemonic: String

    @State private var copied = false

    var body: some View {
        AuthSceneScaffold(
            title: "Save Recovery Phrase",
            subtitle: "Write these words down before you continue.",
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

                Text("A paper backup is safest. If you copy this phrase digitally, ensure the destination is encrypted and offline.")
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

                    Spacer()

                    Button {
                        vault.acceptRecoveryPhrase()
                    } label: {
                        Text("Continue")
                    }
                    .controlSize(.large)
                    .buttonStyle(.borderedProminent)
                    .keyboardShortcut(.defaultAction)
                    .accessibilityHint("Continue after you have saved the recovery phrase.")
                }
            }
        }
    }
}
