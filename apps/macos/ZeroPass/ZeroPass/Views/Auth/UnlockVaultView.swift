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
    @State private var flagsMonitor: Any?

    private enum UnlockMethod: String, CaseIterable, Identifiable {
        case password = "Master Password"
        case recoveryPhrase = "Recovery Phrase"

        var id: String { rawValue }
    }


    var body: some View {
        ZStack {
            ZPTheme.authSceneBackground
                .ignoresSafeArea()

            backgroundMotifs

            VStack(spacing: 0) {
                GeometryReader { proxy in
                    ScrollView(showsIndicators: false) {
                        VStack(spacing: 0) {
                            unlockScene
                                .frame(maxWidth: 420)
                                .frame(maxWidth: .infinity)
                                .offset(x: shakeOffset)

                            Spacer()
                        }
                        .frame(minHeight: proxy.size.height)
                        .padding(.horizontal, ZPTheme.spacing32)
                        .padding(.top, ZPTheme.spacing40)
                        .padding(.bottom, ZPTheme.spacing20)
                    }
                }

                bottomFootnoteBar
            }
        }
        // .toolbar routes items to the NSWindow toolbar even without NavigationStack,
        // which is correct here since ContentView uses mutually exclusive state switching.
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Menu {
                    Button("Choose Different Vault…") {
                        showOpenVaultSheet = true
                    }

                    Divider()

                    Button("Close Vault…", role: .destructive) {
                        showCloseConfirmation = true
                    }
                } label: {
                    Image(systemName: "ellipsis")
                        .font(.system(size: 14, weight: .bold))
                        .foregroundStyle(ZPTheme.textSecondary)
                }
                .menuStyle(.borderlessButton)
                .disabled(isBusy)
                .accessibilityLabel("Vault options")
                .help("Vault options")
            }
        }
        .sheet(isPresented: $showOpenVaultSheet, onDismiss: restoreActiveUnlockFocusIfNeeded) {
            OpenVaultSheet(mode: .replaceCurrent)
                .environmentObject(vault)
        }
        .onAppear {
            updateCapsLock()
            errorMessage = nil
            flagsMonitor = NSEvent.addLocalMonitorForEvents(matching: .flagsChanged) { [self] event in
                capsLockOn = selectedMethod == .password && event.modifierFlags.contains(.capsLock)
                return event
            }
        }
        .onDisappear {
            if let monitor = flagsMonitor {
                NSEvent.removeMonitor(monitor)
                flagsMonitor = nil
            }
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
            vaultContextChip
            unlockPanel
            footerMeta
        }
        .frame(maxWidth: 392)
    }

    private var vaultContextChip: some View {
        HStack(spacing: ZPTheme.spacing8) {
            Image(systemName: "lock.fill")
                .font(.system(size: 11, weight: .bold))
                .foregroundStyle(ZPTheme.accent)
                .frame(width: 14, height: 14)
                .accessibilityHidden(true)

            VStack(alignment: .leading, spacing: 2) {
                Text(vaultSubtitle ?? "Current Vault")
                    .font(.caption.weight(.semibold))
                    .foregroundStyle(ZPTheme.textPrimary)

                if let vaultDetail, !vaultDetail.isEmpty {
                    Text(vaultDetail)
                        .font(.system(size: 10, weight: .medium, design: .monospaced))
                        .foregroundStyle(ZPTheme.textTertiary)
                        .lineLimit(1)
                        .truncationMode(.middle)
                }
            }

            Spacer(minLength: 0)
        }
        .padding(.horizontal, ZPTheme.spacing12)
        .padding(.vertical, ZPTheme.spacing8)
        .frame(maxWidth: 320, alignment: .leading)
        .background(ZPTheme.chipBackground, in: Capsule())
        .overlay(
            Capsule()
                .stroke(ZPTheme.panelBorder, lineWidth: 1)
        )
    }

    private var unlockPanel: some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing16) {
            unlockMethodPicker

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

            .animation(.easeInOut(duration: 0.15), value: selectedMethod)

            if let errorMessage, !errorMessage.isEmpty {
                HStack(spacing: ZPTheme.spacing8) {
                    Image(systemName: "exclamationmark.triangle.fill")
                        .font(.caption2.weight(.semibold))

                    Text(errorMessage)
                        .font(.caption2)
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
                        .font(.caption.weight(.bold))
                }
                .font(.system(size: 13, weight: .semibold))
                .foregroundStyle(.white)
                .frame(maxWidth: .infinity)
                .frame(height: 46)
                .background(
                    RoundedRectangle(cornerRadius: 10, style: .continuous)
                        .fill(isUnlockDisabled ? ZPTheme.unlockButton.opacity(0.45) : ZPTheme.unlockButton)
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
                        .font(.caption.weight(.medium))
                        .foregroundStyle(ZPTheme.textSecondary)
                        .frame(maxWidth: .infinity)
                }
                .buttonStyle(.plain)
                .disabled(isBusy)
                .accessibilityLabel("Unlock with Touch ID")
                .accessibilityIdentifier("unlockVault.touchIDButton")
            }
        }
        .padding(.horizontal, ZPTheme.spacing24)
        .padding(.vertical, ZPTheme.spacing24)
        .frame(maxWidth: 360)
        .background(
            RoundedRectangle(cornerRadius: ZPTheme.radiusPanel, style: .continuous)
                .fill(ZPTheme.panelBackgroundElevated)
        )
        .overlay(
            RoundedRectangle(cornerRadius: ZPTheme.radiusPanel, style: .continuous)
                .stroke(ZPTheme.panelBorderStrong, lineWidth: 1)
        )
        .shadow(color: Color.black.opacity(0.18), radius: ZPTheme.radiusPanel, y: 10)
    }

    private var footerMeta: some View {
        HStack(spacing: ZPTheme.spacing8) {
            unlockStatusBadge("AES-256 Encrypted", dotColor: ZPTheme.success)
            unlockStatusBadge(autoLockBadgeTitle, systemImage: vault.autoLockTimeoutSeconds == 0 ? "lock.open.display" : "timer")

            if canUseBiometrics && selectedMethod == .password {
                unlockStatusBadge("Touch ID Ready", systemImage: "checkmark.circle.fill")
            }
        }
        .frame(maxWidth: .infinity, alignment: .center)
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

    private var bottomFootnoteText: String {
        if selectedMethod == .password {
            return biometricHelperText ?? "Zero-knowledge: your master password never leaves this machine."
        }

        return "Recovery unlock rotates the phrase after use. Save the replacement phrase offline immediately."
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

    private var unlockMethodPicker: some View {
        HStack(spacing: ZPTheme.spacing4) {
            ForEach(UnlockMethod.allCases) { method in
                Button {
                    selectedMethod = method
                } label: {
                    Text(method.rawValue)
                        .font(.system(size: 11, weight: .semibold))
                        .foregroundStyle(selectedMethod == method ? ZPTheme.textPrimary : ZPTheme.textTertiary)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 7)
                        .background(
                            RoundedRectangle(cornerRadius: 8, style: .continuous)
                                .fill(selectedMethod == method ? ZPTheme.chipBackground : .clear)
                        )
                }
                .buttonStyle(.plain)
                .disabled(isBusy)
                .accessibilityLabel(method.rawValue)
                .accessibilityAddTraits(selectedMethod == method ? .isSelected : [])
            }
        }
        .padding(4)
        .background(
            RoundedRectangle(cornerRadius: 10, style: .continuous)
                .fill(ZPTheme.authInsetBackground)
        )
        .overlay(
            RoundedRectangle(cornerRadius: 10, style: .continuous)
                .stroke(ZPTheme.panelBorder.opacity(0.7), lineWidth: 1)
        )
        .accessibilityElement(children: .contain)
        .accessibilityLabel("Unlock method")
        .accessibilityHint("Choose whether to unlock with your master password or recovery phrase")
    }

    private var bottomFootnoteBar: some View {
        HStack {
            Text(bottomFootnoteText)
                .font(.system(size: 10, weight: .medium))
                .foregroundStyle(ZPTheme.textTertiary)
                .multilineTextAlignment(.center)
                .frame(maxWidth: .infinity)
        }
        .padding(.horizontal, ZPTheme.spacing16)
        .padding(.vertical, ZPTheme.spacing8)
        .background(ZPTheme.authSceneBackground.opacity(0.94))
        .overlay(alignment: .top) {
            Rectangle()
                .fill(ZPTheme.separatorSubtle)
                .frame(height: 1)
        }
    }

    @ViewBuilder
    private var backgroundMotifs: some View {
        Canvas { context, size in
            let streaks: [(start: CGPoint, end: CGPoint, width: CGFloat, opacity: Double)] = [
                (CGPoint(x: size.width * 0.15, y: -20), CGPoint(x: size.width * 0.45, y: size.height + 20), 2.5, 0.045),
                (CGPoint(x: size.width * 0.25, y: -20), CGPoint(x: size.width * 0.55, y: size.height + 20), 1.5, 0.035),
                (CGPoint(x: size.width * 0.50, y: -20), CGPoint(x: size.width * 0.80, y: size.height + 20), 3.0, 0.04),
                (CGPoint(x: size.width * 0.70, y: -20), CGPoint(x: size.width * 1.0, y: size.height + 20), 1.8, 0.03),
            ]

            for streak in streaks {
                var path = Path()
                path.move(to: streak.start)
                path.addLine(to: streak.end)
                context.stroke(
                    path,
                    with: .color(.white.opacity(streak.opacity)),
                    lineWidth: streak.width
                )
            }
        }
        .ignoresSafeArea()
        .allowsHitTesting(false)
        .accessibilityHidden(true)
    }

    @ViewBuilder
    private func unlockStatusBadge(_ title: String, systemImage: String? = nil, dotColor: Color? = nil) -> some View {
        HStack(spacing: 6) {
            if let dotColor {
                Circle()
                    .fill(dotColor)
                    .frame(width: 6, height: 6)
            }

            if let systemImage {
                Image(systemName: systemImage)
                    .font(.system(size: 11, weight: .semibold))
                    .foregroundStyle(ZPTheme.textTertiary)
            }

            Text(title.uppercased())
                .font(.system(size: 10, weight: .medium))
                .tracking(0.7)
                .foregroundStyle(ZPTheme.textSecondary)
        }
        .padding(.horizontal, ZPTheme.spacing10)
        .padding(.vertical, 5)
        .background(
            RoundedRectangle(cornerRadius: 8, style: .continuous)
                .fill(ZPTheme.chipBackground)
        )
        .overlay(
            RoundedRectangle(cornerRadius: 8, style: .continuous)
                .stroke(ZPTheme.panelBorder.opacity(0.72), lineWidth: 1)
        )
    }
}
