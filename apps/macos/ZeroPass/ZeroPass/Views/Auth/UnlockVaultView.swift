import SwiftUI

struct UnlockVaultView: View {
    @EnvironmentObject var vault: VaultClient

    @State private var password: String = ""
    @State private var mnemonic: String = ""
    @State private var useRecovery = false
    @State private var isBusy = false
    @State private var showError = false
    @State private var errorMessage = ""
    @State private var iconAppeared = false
    @State private var shakeOffset: CGFloat = 0

    private enum Field { case password, mnemonic }
    @FocusState private var focusedField: Field?

    var body: some View {
        VStack(spacing: 0) {
            Spacer()

            VStack(spacing: 24) {
                Image(systemName: "lock.fill")
                    .font(.system(size: 48))
                    .foregroundStyle(.tint)
                    .symbolEffect(.appear, isActive: iconAppeared)
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
                        HStack(spacing: 8) {
                            SecureField("Master Password", text: $password)
                                .textFieldStyle(.roundedBorder)
                                .onSubmit { unlock() }
                                .focused($focusedField, equals: .password)
                                .accessibilityLabel("Master password")

                            if vault.biometricUnlockEnabled, vault.isBiometricAvailable {
                                Button {
                                    unlockWithBiometrics()
                                } label: {
                                    Image(systemName: "touchid")
                                        .font(.title2)
                                }
                                .buttonStyle(.borderless)
                                .disabled(isBusy)
                                .help("Unlock with Touch ID")
                                .accessibilityLabel("Unlock with Touch ID")
                            }
                        }
                    }

                    Toggle("Use recovery phrase", isOn: $useRecovery)
                        .font(.callout)
                        .toggleStyle(.checkbox)
                }
                .frame(maxWidth: 300)

                Button {
                    unlock()
                } label: {
                    Text("Unlock")
                        .frame(maxWidth: 200)
                }
                .controlSize(.large)
                .buttonStyle(.borderedProminent)
                .keyboardShortcut(.defaultAction)
                .disabled(isBusy || (useRecovery ? mnemonic.isEmpty : password.isEmpty))
                .accessibilityLabel("Unlock vault")
            }
            .offset(x: shakeOffset)

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
        .frame(minWidth: 560, minHeight: 420)
        .onAppear {
            iconAppeared = true
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

    private func shake() {
        withAnimation(.spring(response: 0.1, dampingFraction: 0.3)) {
            shakeOffset = 10
        }
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.1) {
            withAnimation(.spring(response: 0.1, dampingFraction: 0.3)) {
                shakeOffset = -8
            }
        }
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.2) {
            withAnimation(.spring(response: 0.15, dampingFraction: 0.5)) {
                shakeOffset = 0
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
                shake()
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
