import SwiftUI

struct UnlockPasswordSection: View {
    @Binding var password: String
    @Binding var showPassword: Bool

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
                .font(.caption)
                .fontWeight(.medium)
                .foregroundStyle(ZPTheme.textSecondary)

            HStack(spacing: ZPTheme.spacing8) {
                Group {
                    if showPassword {
                        TextField("Enter your master password", text: $password)
                            .textFieldStyle(.plain)
                    } else {
                        SecureField("Enter your master password", text: $password)
                            .textFieldStyle(.plain)
                    }
                }
                .font(.body)
                .autocorrectionDisabled()
                .focused($isPasswordFocused)
                .onSubmit(onSubmit)
                .accessibilityLabel("Master password")

                Button {
                    showPassword.toggle()
                } label: {
                    Image(systemName: showPassword ? "eye.slash" : "eye")
                        .font(.body.weight(.semibold))
                        .foregroundStyle(ZPTheme.textSecondary)
                }
                .buttonStyle(.borderless)
                .accessibilityLabel(showPassword ? "Hide password" : "Show password")
                .help(showPassword ? "Hide password" : "Show password")
            }
            .padding(.horizontal, ZPTheme.spacing12)
            .padding(.vertical, ZPTheme.spacing10)
            .frame(minHeight: ZPTheme.authFieldHeight)
            .background(ZPTheme.authInsetBackground, in: RoundedRectangle(cornerRadius: ZPTheme.radiusLarge, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: ZPTheme.radiusLarge, style: .continuous)
                    .stroke(borderColor, lineWidth: borderWidth)
            )
            .shadow(color: ZPTheme.fieldShadow, radius: 2, y: 1)

            if capsLockOn {
                Label("Caps Lock is on", systemImage: "capslock.fill")
                    .font(.caption)
                    .foregroundStyle(ZPTheme.warning)
            }

            if canUseBiometrics {
                HStack(alignment: .top, spacing: ZPTheme.spacing12) {
                    Button {
                        onUnlockWithBiometrics()
                    } label: {
                        Label("Unlock with Touch ID", systemImage: "touchid")
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
            }
        }
        .onAppear {
            DispatchQueue.main.async {
                isPasswordFocused = true
            }
        }
    }

    private var borderColor: Color {
        if showErrorHighlight {
            return ZPTheme.destructive
        }
        return isPasswordFocused ? ZPTheme.info : ZPTheme.authInsetBorder
    }

    private var borderWidth: CGFloat {
        (showErrorHighlight || isPasswordFocused) ? 1.5 : 1
    }
}
