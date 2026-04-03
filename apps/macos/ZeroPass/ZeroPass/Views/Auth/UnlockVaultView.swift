import SwiftUI

struct UnlockVaultView: View {
    @EnvironmentObject var vault: VaultClient
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    @State private var password: String = ""
    @State private var mnemonic: String = ""
    @State private var selectedMethod: UnlockMethod = .password
    @State private var showPassword = false
    @State private var isBusy = false
    @State private var errorMessage: String?
    @State private var shakeOffset: CGFloat = 0
    @State private var showErrorHighlight = false
    @State private var capsLockOn = false
    @State private var showCloseConfirmation = false
    @State private var showOpenVaultSheet = false

    private enum UnlockMethod: String, CaseIterable, Identifiable {
        case password = "Password"
        case recoveryPhrase = "Recovery Phrase"

        var id: String { rawValue }
    }


    var body: some View {
        AuthSceneScaffold(
            title: "Unlock Vault",
            subtitle: vaultSubtitle,
            detail: vaultDetail,
            detailSymbolName: vaultDetail == nil ? nil : "folder",
            symbolName: "lock.shield.fill",
            accessory: {
                vaultMenu
            }
        ) {
            VStack(alignment: .leading, spacing: ZPTheme.spacing16) {
                Picker("Unlock method", selection: $selectedMethod) {
                    ForEach(UnlockMethod.allCases) { method in
                        Text(method.rawValue)
                            .tag(method)
                    }
                }
                .pickerStyle(.segmented)
                .disabled(isBusy)
                .accessibilityHint("Choose whether to unlock with your master password or recovery phrase")

                Group {
                    if selectedMethod == .password {
                        UnlockPasswordSection(
                            password: $password,
                            showPassword: $showPassword,
                            isBusy: isBusy,
                            showErrorHighlight: showErrorHighlight,
                            capsLockOn: capsLockOn,
                            canUseBiometrics: canUseBiometrics,
                            biometricHelperText: biometricHelperText,
                            onSubmit: unlock,
                            onUnlockWithBiometrics: unlockWithBiometrics
                        )
                    } else {
                        UnlockRecoverySection(
                            mnemonic: $mnemonic,
                            showErrorHighlight: showErrorHighlight
                        )
                    }
                }

                if let errorMessage, !errorMessage.isEmpty {
                    AuthMessageView(
                        text: errorMessage,
                        systemImage: "exclamationmark.triangle.fill",
                        tone: .error
                    )
                }

                Button {
                    unlock()
                } label: {
                    HStack(spacing: ZPTheme.spacing8) {
                        if isBusy {
                            ProgressView()
                                .controlSize(.small)
                        }
                        Text(isBusy ? "Unlocking…" : "Unlock")
                            .frame(maxWidth: .infinity)
                    }
                }
                .controlSize(.large)
                .buttonStyle(.borderedProminent)
                .keyboardShortcut(.defaultAction)
                .disabled(isUnlockDisabled)
                .accessibilityLabel("Unlock vault")

                AuthSupportingNoteView(
                    text: "Encrypted locally. ZeroPass never sends your master password anywhere.",
                    systemImage: "checkmark.shield"
                )
            }
        }
        .offset(x: shakeOffset)
        .sheet(isPresented: $showOpenVaultSheet) {
            OpenVaultSheet(mode: .replaceCurrent)
                .environmentObject(vault)
        }
        .onAppear {
            updateCapsLock()
            errorMessage = nil
        }
        .onChange(of: selectedMethod) { _, newValue in
            errorMessage = nil
            showErrorHighlight = false
            capsLockOn = newValue == .password && NSEvent.modifierFlags.contains(.capsLock)
        }
        .onReceive(NotificationCenter.default.publisher(for: NSApplication.didBecomeActiveNotification)) { _ in
            updateCapsLock()
        }
        .confirmationDialog(
            "Close Vault?",
            isPresented: $showCloseConfirmation,
            titleVisibility: .visible
        ) {
            Button("Close Vault", role: .destructive) {
                Task { await vault.closeVault() }
            }
            Button("Cancel", role: .cancel) {}
        } message: {
            Text("You can reopen this vault later from the welcome screen.")
        }
    }

    private var vaultMenu: some View {
        HStack(spacing: ZPTheme.spacing8) {
            Button {
                closeWindow()
            } label: {
                Image(systemName: "xmark")
                    .font(.caption.weight(.semibold))
                    .foregroundStyle(ZPTheme.textSecondary)
                    .frame(width: 28, height: 28)
            }
            .buttonStyle(.plain)
            .background(.quaternary.opacity(0.7), in: RoundedRectangle(cornerRadius: ZPTheme.radiusMedium, style: .continuous))
            .help("Close window")
            .accessibilityLabel("Close window")

            Menu {
                Button("Choose Different Vault…") {
                    showOpenVaultSheet = true
                }

                Divider()

                Button("Close Vault…", role: .destructive) {
                    showCloseConfirmation = true
                }
            } label: {
                Label("Vault", systemImage: "ellipsis.circle")
                    .font(.callout)
                    .foregroundStyle(ZPTheme.textSecondary)
            }
            .disabled(isBusy)
            .accessibilityLabel("Vault actions")
        }
    }

    private var vaultSubtitle: String? {
        let name = vault.vaultName.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !name.isEmpty else { return nil }
        return name
    }

    private var vaultDetail: String? {
        let path = vault.vaultPathDisplay.trimmingCharacters(in: .whitespacesAndNewlines)
        return path == "—" ? nil : path
    }

    private var canUseBiometrics: Bool {
        vault.biometricUnlockEnabled && vault.isBiometricAvailable
    }

    private var biometricHelperText: String? {
        guard !canUseBiometrics, vault.isBiometricAvailable else { return nil }
        return "Touch ID is available on this Mac. Enable it in Security settings after unlocking once."
    }

    private var isUnlockDisabled: Bool {
        if isBusy {
            return true
        }

        switch selectedMethod {
        case .password:
            return password.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
        case .recoveryPhrase:
            return RecoveryPhraseSupport.normalizedText(from: mnemonic).isEmpty
        }
    }

    private func updateCapsLock() {
        capsLockOn = selectedMethod == .password && NSEvent.modifierFlags.contains(.capsLock)
    }

    private func shake() {
        guard !reduceMotion else {
            withAnimation(.easeInOut(duration: 0.3)) { showErrorHighlight = true }
            DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) {
                withAnimation(.easeInOut(duration: 0.3)) { showErrorHighlight = false }
            }
            return
        }

        showErrorHighlight = true
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
        DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) {
            withAnimation(.easeInOut(duration: 0.3)) { showErrorHighlight = false }
        }
    }

    private func unlock() {
        isBusy = true
        errorMessage = nil
        vault.lastError = nil
        showErrorHighlight = false

        Task {
            defer { isBusy = false }

            do {
                if selectedMethod == .password {
                    try await vault.unlock(masterPassword: password)
                } else {
                    try await vault.unlockWithRecovery(
                        mnemonic: RecoveryPhraseSupport.normalizedText(from: mnemonic)
                    )
                }
            } catch {
                presentError(error.localizedDescription)
            }
        }
    }

    private func unlockWithBiometrics() {
        isBusy = true
        errorMessage = nil
        vault.lastError = nil
        showErrorHighlight = false

        Task {
            defer { isBusy = false }

            do {
                try await vault.unlockWithBiometrics()
            } catch {
                presentError(error.localizedDescription)
            }
        }
    }

    private func presentError(_ message: String) {
        errorMessage = message
        shake()
    }

    private func closeWindow() {
        NSApp.keyWindow?.close()
    }
}
