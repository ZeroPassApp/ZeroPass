import AppKit
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
        case password = "Master Password"
        case recoveryPhrase = "Recovery Phrase"

        var id: String { rawValue }
    }


    var body: some View {
        ZStack {
            ZPTheme.authSceneBackground
                .ignoresSafeArea()

            GeometryReader { proxy in
                ScrollView {
                    VStack(spacing: 0) {
                        Spacer(minLength: 0)

                        unlockScene
                            .frame(maxWidth: 460)
                            .frame(maxWidth: .infinity)

                        Spacer(minLength: 0)
                    }
                    .frame(
                        minHeight: max(
                            proxy.size.height - (ZPTheme.spacing32 * 2),
                            CGFloat.zero
                        )
                    )
                    .padding(.horizontal, ZPTheme.spacing32)
                    .padding(.vertical, ZPTheme.spacing32)
                }
            }
        }
        .offset(x: shakeOffset)
        .sheet(isPresented: $showOpenVaultSheet, onDismiss: restoreActiveUnlockFocusIfNeeded) {
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

    private var unlockScene: some View {
        VStack(spacing: ZPTheme.spacing20) {
            topBar
            vaultContextChip
            unlockPanel
            footerMeta
        }
        .frame(maxWidth: 430)
    }

    private var topBar: some View {
        HStack {
            Spacer(minLength: 0)
            vaultMenu
        }
        .frame(maxWidth: .infinity)
        .padding(.trailing, ZPTheme.spacing4)
    }

    private var vaultContextChip: some View {
        HStack(spacing: ZPTheme.spacing10) {
            ZStack {
                Circle()
                    .fill(ZPTheme.accentSoft)

                Image(systemName: "lock.fill")
                    .font(.system(size: 10, weight: .bold))
                    .foregroundStyle(ZPTheme.accent)
            }
            .frame(width: 24, height: 24)

            VStack(alignment: .leading, spacing: 2) {
                Text(vaultSubtitle ?? "Current Vault")
                    .font(.callout.weight(.semibold))
                    .foregroundStyle(ZPTheme.textPrimary)

                if let vaultDetail, !vaultDetail.isEmpty {
                    Text(vaultDetail)
                        .font(.caption2)
                        .foregroundStyle(ZPTheme.textTertiary)
                        .lineLimit(1)
                        .truncationMode(.middle)
                }
            }

            Spacer(minLength: 0)
        }
        .padding(.horizontal, ZPTheme.spacing12)
        .padding(.vertical, ZPTheme.spacing10)
        .frame(maxWidth: 340, alignment: .leading)
        .background(ZPTheme.chipBackground, in: RoundedRectangle(cornerRadius: ZPTheme.radiusLarge, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: ZPTheme.radiusLarge, style: .continuous)
                .stroke(ZPTheme.panelBorder, lineWidth: 1)
        )
    }

    private var unlockPanel: some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing14) {
            Picker("Unlock method", selection: $selectedMethod) {
                ForEach(UnlockMethod.allCases) { method in
                    Text(method.rawValue)
                        .tag(method)
                }
            }
            .pickerStyle(.segmented)
            .labelsHidden()
            .controlSize(.small)
            .disabled(isBusy)
            .accessibilityHint("Choose whether to unlock with your master password or recovery phrase")

            Group {
                if selectedMethod == .password {
                    UnlockPasswordSection(
                        password: $password,
                        showPassword: $showPassword,
                        focusRequestID: vault.unlockPasswordFocusRequestID,
                        isBusy: isBusy,
                        showErrorHighlight: showErrorHighlight,
                        capsLockOn: capsLockOn,
                        onSubmit: unlock
                    )
                } else {
                    UnlockRecoverySection(
                        mnemonic: $mnemonic,
                        focusRequestID: vault.unlockRecoveryFocusRequestID,
                        showErrorHighlight: showErrorHighlight
                    )
                }
            }

            if let errorMessage, !errorMessage.isEmpty {
                HStack(spacing: ZPTheme.spacing8) {
                    Image(systemName: "exclamationmark.triangle.fill")
                        .font(.caption.weight(.semibold))

                    Text(errorMessage)
                        .font(.caption)
                        .fixedSize(horizontal: false, vertical: true)
                }
                .foregroundStyle(ZPTheme.destructive)
                .accessibilityElement(children: .combine)
            }

            Button {
                unlock()
            } label: {
                HStack(spacing: ZPTheme.spacing8) {
                    if isBusy {
                        ProgressView()
                            .controlSize(.small)
                            .tint(.white)
                    }

                    Text(isBusy ? "Unlocking…" : "Unlock")

                    Image(systemName: "arrow.right")
                        .font(.footnote.weight(.bold))
                }
                .font(.callout.weight(.semibold))
                .foregroundStyle(.white)
                .frame(maxWidth: .infinity)
                .frame(height: 38)
                .background(
                    RoundedRectangle(cornerRadius: ZPTheme.radiusMedium, style: .continuous)
                        .fill(isUnlockDisabled ? ZPTheme.accent.opacity(0.45) : ZPTheme.accent)
                )
            }
            .buttonStyle(.plain)
            .keyboardShortcut(.defaultAction)
            .disabled(isUnlockDisabled)
            .accessibilityLabel("Unlock vault")
            .accessibilityHint("Unlocks the selected vault")
            .accessibilityIdentifier("unlockVault.submitButton")

            if canUseBiometrics && selectedMethod == .password {
                Button {
                    unlockWithBiometrics()
                } label: {
                    Label("Use Touch ID", systemImage: "touchid")
                        .font(.caption.weight(.semibold))
                        .foregroundStyle(ZPTheme.textSecondary)
                        .frame(maxWidth: .infinity)
                }
                .buttonStyle(.plain)
                .disabled(isBusy)
                .accessibilityLabel("Unlock with Touch ID")
                .accessibilityIdentifier("unlockVault.touchIDButton")
            }
        }
        .padding(ZPTheme.spacing20)
        .frame(maxWidth: 372)
        .zpSurface(.elevated, radius: 20, shadow: false)
    }

    private var footerMeta: some View {
        VStack(spacing: ZPTheme.spacing12) {
            HStack(spacing: ZPTheme.spacing8) {
                unlockPill("AES-256 Encrypted", systemImage: "lock.shield")
                unlockPill(autoLockBadgeTitle, systemImage: vault.autoLockTimeoutSeconds == 0 ? "lock.open.display" : "timer")

                if canUseBiometrics && selectedMethod == .password {
                    unlockPill("Touch ID Ready", systemImage: "touchid")
                }
            }
            .frame(maxWidth: .infinity, alignment: .center)

            HStack(alignment: .top, spacing: ZPTheme.spacing6) {
                Spacer(minLength: 0)

                Image(systemName: "checkmark.circle.fill")
                    .font(.caption.weight(.semibold))
                    .foregroundStyle(ZPTheme.textTertiary)
                    .padding(.top, 1)

                Text(selectedMethod == .password
                     ? "Zero-knowledge: your master password never leaves this machine."
                     : "Recovery unlock rotates the phrase after use. Save the replacement phrase offline immediately.")
                    .font(.caption2)
                    .foregroundStyle(ZPTheme.textSecondary)
                    .multilineTextAlignment(.center)
                    .fixedSize(horizontal: false, vertical: true)

                Spacer(minLength: 0)
            }

            if selectedMethod == .password,
               let biometricHelperText,
               !biometricHelperText.isEmpty,
               !canUseBiometrics {
                Text(biometricHelperText)
                    .font(.caption2)
                    .foregroundStyle(ZPTheme.textTertiary)
                    .multilineTextAlignment(.center)
                    .fixedSize(horizontal: false, vertical: true)
            }
        }
    }

    private var vaultMenu: some View {
        Menu {
            Button("Choose Different Vault…") {
                showOpenVaultSheet = true
            }

            Divider()

            Button("Close Vault…", role: .destructive) {
                showCloseConfirmation = true
            }
        } label: {
            HStack(spacing: 2) {
                Image(systemName: "ellipsis")
                    .font(.system(size: 14, weight: .bold))

                Image(systemName: "chevron.down")
                    .font(.system(size: 10, weight: .semibold))
            }
            .foregroundStyle(ZPTheme.textSecondary)
            .frame(width: 34, height: 28)
            .contentShape(Rectangle())
        }
        .menuStyle(BorderlessButtonMenuStyle())
        .disabled(isBusy)
        .accessibilityLabel("Vault options")
        .accessibilityHint("Open actions for the current vault")
        .help("Vault options")
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

    private var autoLockSummary: String {
        guard vault.autoLockTimeoutSeconds > 0 else {
            return "Manual lock"
        }

        let minutes = vault.autoLockTimeoutSeconds / 60
        if minutes > 0 {
            return "Auto-lock in \(minutes)m"
        }
        return "Auto-lock in \(vault.autoLockTimeoutSeconds)s"
    }

    private var autoLockBadgeTitle: String {
        guard vault.autoLockTimeoutSeconds > 0 else {
            return "Manual Lock"
        }

        let minutes = vault.autoLockTimeoutSeconds / 60
        if minutes > 0 {
            return "Auto-lock: \(minutes)m"
        }

        return "Auto-lock: \(vault.autoLockTimeoutSeconds)s"
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
        guard !isBusy else { return }

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
        guard !isBusy else { return }

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
        announceError(message)
        shake()
    }

    private func announceError(_ message: String) {
        guard let application = NSApp else { return }

        NSAccessibility.post(
            element: application,
            notification: .announcementRequested,
            userInfo: [
                .announcement: message,
                .priority: NSAccessibilityPriorityLevel.high.rawValue
            ]
        )
    }

    private func restoreActiveUnlockFocusIfNeeded() {
        switch selectedMethod {
        case .password:
            vault.requestUnlockPasswordFocus()
        case .recoveryPhrase:
            vault.requestUnlockRecoveryFocus()
        }
    }

    @ViewBuilder
    private func unlockPill(_ title: String, systemImage: String) -> some View {
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
