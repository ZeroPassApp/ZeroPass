import SwiftUI

struct RecoveryPhraseView: View {
    @EnvironmentObject var vault: VaultClient
    let mnemonic: String

    @State private var copied = false

    private var words: [String] {
        mnemonic.split(separator: " ").map(String.init)
    }

    private let columns = Array(repeating: GridItem(.flexible(), spacing: 12), count: 4)

    var body: some View {
        VStack(spacing: 0) {
            Spacer()

            VStack(spacing: 20) {
                Image(systemName: "key.horizontal.fill")
                    .font(.system(size: 40))
                    .foregroundStyle(.orange)

                VStack(spacing: 6) {
                    Text("Recovery Phrase")
                        .font(.title2)
                        .bold()

                    Text("Write these words down in order and store them somewhere safe.\nThis phrase can unlock your vault if you forget your password.")
                        .font(.callout)
                        .foregroundStyle(.secondary)
                        .multilineTextAlignment(.center)
                        .fixedSize(horizontal: false, vertical: true)
                }

                LazyVGrid(columns: columns, spacing: 10) {
                    ForEach(Array(words.enumerated()), id: \.offset) { index, word in
                        HStack(spacing: 4) {
                            Text("\(index + 1).")
                                .font(.caption)
                                .foregroundStyle(.tertiary)
                                .frame(width: 20, alignment: .trailing)

                            Text(word)
                                .font(.system(.body, design: .monospaced))
                        }
                        .padding(.vertical, 6)
                        .padding(.horizontal, 8)
                        .frame(maxWidth: .infinity)
                        .background(.quaternary)
                        .clipShape(RoundedRectangle(cornerRadius: 6, style: .continuous))
                    }
                }
                .padding(.horizontal, 4)
                .textSelection(.enabled)

                HStack(spacing: 12) {
                    Button {
                        let secs = vault.clipboardAutoClearEnabled ? vault.clipboardAutoClearSeconds : 0
                        ClipboardService.shared.copySensitive(mnemonic, clearAfterSeconds: secs)
                        copied = true
                        DispatchQueue.main.asyncAfter(deadline: .now() + 2) { copied = false }
                    } label: {
                        Label(copied ? "Copied!" : "Copy", systemImage: copied ? "checkmark" : "doc.on.doc")
                    }

                    Spacer()

                    Button {
                        vault.acceptRecoveryPhrase()
                    } label: {
                        Text("I Saved My Recovery Phrase")
                    }
                    .controlSize(.large)
                    .keyboardShortcut(.defaultAction)
                }
            }
            .frame(maxWidth: 520)

            Spacer()
        }
        .padding(24)
        .frame(minWidth: 580, minHeight: 420)
    }
}
