import SwiftUI

struct UnlockVaultView: View {
    @EnvironmentObject var vault: VaultClient

    @State private var password: String = ""
    @State private var mnemonic: String = ""
    @State private var useRecovery = false
    @State private var isBusy = false

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Unlock Vault")
                .font(.title2)
                .bold()

            Toggle("Use recovery phrase", isOn: $useRecovery)

            if useRecovery {
                TextField("Recovery phrase", text: $mnemonic)
            } else {
                SecureField("Master Password", text: $password)
            }

            if let err = vault.lastError {
                Text(err)
                    .foregroundStyle(.red)
                    .textSelection(.enabled)
            }

            HStack {
                Button("Close Vault") {
                    Task { await vault.closeVault() }
                }
                .disabled(isBusy)

                Spacer()

                if !useRecovery, vault.biometricUnlockEnabled, vault.isBiometricAvailable {
                    Button("Unlock with Touch ID") { unlockWithBiometrics() }
                        .disabled(isBusy)
                }

                Button("Unlock") { unlock() }
                    .disabled(isBusy || (useRecovery ? mnemonic.isEmpty : password.isEmpty))
                    .keyboardShortcut(.defaultAction)
            }
        }
        .padding(24)
        .frame(minWidth: 520, minHeight: 240)
    }

    private func unlock() {
        isBusy = true
        vault.lastError = nil

        Task {
            defer { isBusy = false }
            do {
                if useRecovery {
                    try await vault.unlockWithRecovery(mnemonic: mnemonic)
                } else {
                    try await vault.unlock(masterPassword: password)
                }
            } catch {
                vault.lastError = error.localizedDescription
            }
        }
    }

    private func unlockWithBiometrics() {
        isBusy = true
        vault.lastError = nil

        Task {
            defer { isBusy = false }
            do {
                try await vault.unlockWithBiometrics()
            } catch {
                vault.lastError = error.localizedDescription
            }
        }
    }
}
