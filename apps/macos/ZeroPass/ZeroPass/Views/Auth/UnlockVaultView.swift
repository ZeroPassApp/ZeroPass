import SwiftUI

struct UnlockVaultView: View {
    @EnvironmentObject var vault: VaultClient

    @State private var password: String = ""
    @State private var mnemonic: String = ""
    @State private var useRecovery = false
    @State private var isBusy = false
    @State private var showError = false
    @State private var errorMessage = ""

    private enum Field { case password, mnemonic }
    @FocusState private var focusedField: Field?

    var body: some View {
        VStack(spacing: 0) {
            Spacer()

            VStack(spacing: 20) {
                Image(systemName: "lock.fill")
                    .font(.system(size: 48))
                    .foregroundStyle(.secondary)
                    .accessibilityHidden(true)

                VStack(spacing: 4) {
                    Text("Unlock Vault")
                        .font(.title2)
                        .bold()

                    Text(vault.vaultName)
                        .font(.subheadline)
                        .foregroundStyle(.tertiary)
                }

                VStack(spacing: 12) {
                    if useRecovery {
                        TextField("Recovery phrase", text: $mnemonic)
                            .textFieldStyle(.roundedBorder)
                            .focused($focusedField, equals: .mnemonic)
                            .accessibilityLabel("Recovery phrase")
                    } else {
                        SecureField("Master Password", text: $password)
                            .textFieldStyle(.roundedBorder)
                            .onSubmit { unlock() }
                            .focused($focusedField, equals: .password)
                            .accessibilityLabel("Master password")
                    }

                    Toggle("Use recovery phrase", isOn: $useRecovery)
                        .font(.callout)
                        .toggleStyle(.checkbox)
                }
                .frame(maxWidth: 280)

                VStack(spacing: 8) {
                    Button {
                        unlock()
                    } label: {
                        Text("Unlock")
                            .frame(maxWidth: 200)
                    }
                    .controlSize(.large)
                    .keyboardShortcut(.defaultAction)
                    .disabled(isBusy || (useRecovery ? mnemonic.isEmpty : password.isEmpty))
                    .accessibilityLabel("Unlock vault")

                    if !useRecovery, vault.biometricUnlockEnabled, vault.isBiometricAvailable {
                        Button {
                            unlockWithBiometrics()
                        } label: {
                            Label("Unlock with Touch ID", systemImage: "touchid")
                                .frame(maxWidth: 200)
                        }
                        .controlSize(.large)
                        .disabled(isBusy)
                        .accessibilityLabel("Unlock with Touch ID")
                    }
                }
            }

            Spacer()

            Button("Close Vault") {
                Task { await vault.closeVault() }
            }
            .font(.callout)
            .foregroundStyle(.secondary)
            .buttonStyle(.plain)
            .disabled(isBusy)
            .padding(.bottom, 16)
            .accessibilityLabel("Close vault")
        }
        .frame(minWidth: 480, minHeight: 360)
        .onAppear {
            focusedField = useRecovery ? .mnemonic : .password
        }
        .onChange(of: useRecovery) { _, newValue in
            focusedField = newValue ? .mnemonic : .password
        }
        .alert("Unlock Failed", isPresented: $showError) {
            Button("OK", role: .cancel) {}
        } message: {
            Text(errorMessage)
        }
        .onChange(of: vault.lastError) { _, newValue in
            if let error = newValue, !error.isEmpty {
                errorMessage = error
                showError = true
                vault.lastError = nil
            }
        }
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
                errorMessage = error.localizedDescription
                showError = true
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
                errorMessage = error.localizedDescription
                showError = true
            }
        }
    }
}
