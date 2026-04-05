import SwiftUI

struct UnlockPasswordSection: View {
    @Binding var password: String
    @Binding var showPassword: Bool

    let focusRequestID: UUID

    let isBusy: Bool
    let showErrorHighlight: Bool
    let capsLockOn: Bool
    let canUseBiometrics: Bool
    let biometricHelperText: String?
    let onSubmit: () -> Void
    let onUnlockWithBiometrics: () -> Void

    @FocusState private var isPasswordFocused: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing12) {
            Text("Master Password")
                .font(.callout)
                .fontWeight(.medium)
                .foregroundStyle(showErrorHighlight ? ZPTheme.destructive : ZPTheme.textSecondary)

            Group {
                if showPassword {
                    TextField("Enter your master password", text: $password)
                } else {
                    SecureField("Enter your master password", text: $password)
                }
            }
            .textFieldStyle(.plain)
            .font(.body.weight(.medium))
            .autocorrectionDisabled()
            .textContentType(.password)
            .privacySensitive()
            .focused($isPasswordFocused)
            .onSubmit(onSubmit)
            .accessibilityLabel("Master password")
            .accessibilityHint("Enter the master password for this vault")
            .padding(.horizontal, ZPTheme.spacing14)
            .padding(.vertical, ZPTheme.spacing12)
            .frame(minHeight: ZPTheme.authFieldHeight)
            .zpSurface(showErrorHighlight ? .accent : .inset, radius: ZPTheme.radiusLarge, shadow: false)

            Toggle("Show Password", isOn: $showPassword)
                .disabled(isBusy)
                .accessibilityLabel("Show password")
                .accessibilityHint("Reveals the characters in your master password field")
                .toggleStyle(.checkbox)

            if capsLockOn {
                Label("Caps Lock is on", systemImage: "capslock.fill")
                    .font(.caption)
                    .foregroundStyle(ZPTheme.warning)
            }

            if canUseBiometrics {
                VStack(alignment: .leading, spacing: ZPTheme.spacing8) {
                    Button {
                        onUnlockWithBiometrics()
                    } label: {
                        Label("Use Touch ID", systemImage: "touchid")
                    }
                    .buttonStyle(.bordered)
                    .disabled(isBusy)
                    .accessibilityLabel("Unlock with Touch ID")

                    Text("Use the saved biometric key for this vault.")
                        .font(.caption)
                        .foregroundStyle(ZPTheme.textSecondary)
                        .fixedSize(horizontal: false, vertical: true)
                }
            } else if let biometricHelperText, !biometricHelperText.isEmpty {
                Label(biometricHelperText, systemImage: "touchid")
                    .font(.caption)
                    .foregroundStyle(ZPTheme.textSecondary)
                    .fixedSize(horizontal: false, vertical: true)
                    .accessibilityElement(children: .combine)
            }
        }
        .onAppear {
            DispatchQueue.main.async {
                isPasswordFocused = true
            }
        }
        .onChange(of: focusRequestID) { _, _ in
            DispatchQueue.main.async {
                isPasswordFocused = true
            }
        }
    }
}
