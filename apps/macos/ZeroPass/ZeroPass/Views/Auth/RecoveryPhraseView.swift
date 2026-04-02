import SwiftUI

struct RecoveryPhraseView: View {
    @EnvironmentObject var vault: VaultClient
    let mnemonic: String

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Recovery Phrase")
                .font(.title2)
                .bold()

            Text("Write this down and store it somewhere safe. It can unlock your vault if you forget your password.")
                .foregroundStyle(.secondary)

            Text(mnemonic)
                .font(.system(.body, design: .monospaced))
                .padding(12)
                .frame(maxWidth: .infinity, alignment: .leading)
                .background(.quaternary)
                .cornerRadius(8)
                .textSelection(.enabled)

            HStack {
                Button("Copy") {
                    let secs = vault.clipboardAutoClearEnabled ? vault.clipboardAutoClearSeconds : 0
                    ClipboardService.shared.copySensitive(mnemonic, clearAfterSeconds: secs)
                }

                Spacer()

                Button("I saved it") {
                    vault.acceptRecoveryPhrase()
                }
                .keyboardShortcut(.defaultAction)
            }
        }
        .padding(24)
        .frame(minWidth: 560, minHeight: 320)
    }
}
