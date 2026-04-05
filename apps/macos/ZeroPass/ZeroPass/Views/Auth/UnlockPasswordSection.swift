import SwiftUI

struct UnlockPasswordSection: View {
    @Binding var password: String
    @Binding var showPassword: Bool

    let focusRequestID: UUID

    let isBusy: Bool
    let showErrorHighlight: Bool
    let capsLockOn: Bool
    let onSubmit: () -> Void

    @FocusState private var isPasswordFocused: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: ZPTheme.spacing8) {
            HStack(spacing: ZPTheme.spacing10) {
                Group {
                    if showPassword {
                        TextField("Enter Master Password", text: $password)
                            .accessibilityIdentifier("unlockVault.passwordField")
                    } else {
                        SecureField("Enter Master Password", text: $password)
                            .accessibilityIdentifier("unlockVault.passwordField")
                    }
                }
                .textFieldStyle(.plain)
                .font(.callout.weight(.medium))
                .autocorrectionDisabled()
                .textContentType(.password)
                .privacySensitive()
                .focused($isPasswordFocused)
                .onSubmit(onSubmit)
                .accessibilityLabel("Master password")
                .accessibilityHint("Enter the master password for this vault")

                Button {
                    showPassword.toggle()
                } label: {
                    Image(systemName: showPassword ? "eye.slash" : "eye")
                        .font(.system(size: 13, weight: .semibold))
                        .foregroundStyle(ZPTheme.accent)
                        .frame(width: 24, height: 24)
                }
                .buttonStyle(.borderless)
                .disabled(isBusy)
                .accessibilityLabel("Password visibility")
                .accessibilityValue(showPassword ? "Visible" : "Hidden")
                .accessibilityHint(showPassword ? "Hide the master password" : "Show the master password")
            }
            .padding(.horizontal, ZPTheme.spacing12)
            .padding(.vertical, ZPTheme.spacing10)
            .frame(minHeight: 40)
            .zpSurface(.inset, radius: ZPTheme.radiusMedium, shadow: false)
            .overlay(
                RoundedRectangle(cornerRadius: ZPTheme.radiusMedium, style: .continuous)
                    .stroke(showErrorHighlight ? ZPTheme.destructive : Color.clear, lineWidth: 1.5)
            )

            if capsLockOn {
                Label("Caps Lock is on", systemImage: "capslock.fill")
                    .font(.caption2)
                    .foregroundStyle(ZPTheme.warning)
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
