import SwiftUI

struct RecoveryPhraseView: View {
    @EnvironmentObject var vault: VaultClient
    let mnemonic: String

    @State private var copied = false

    var body: some View {
        AuthSceneScaffold(
            title: "Recovery Phrase",
            subtitle: "Save these words in order before continuing.",
            detail: "This phrase can unlock your vault if you ever lose your master password.",
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
                        Text("I Saved My Recovery Phrase")
                    }
                    .controlSize(.large)
                    .buttonStyle(.borderedProminent)
                    .keyboardShortcut(.defaultAction)
                }
            }
        }
    }
}
