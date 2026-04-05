import SwiftUI

struct ToastOverlay: ViewModifier {
    @Binding var isShowing: Bool
    let message: String
    let icon: String
    @State private var dismissTask: Task<Void, Never>?

    func body(content: Content) -> some View {
        content.overlay(alignment: .bottom) {
            if isShowing {
                HStack(spacing: ZPTheme.spacing8) {
                    Image(systemName: icon)
                        .font(.system(size: 12, weight: .semibold))
                        .foregroundStyle(ZPTheme.success)
                    Text(message)
                        .font(.system(size: 12, weight: .semibold))
                        .foregroundStyle(ZPTheme.textPrimary)
                }
                .padding(.horizontal, ZPTheme.spacing16)
                .padding(.vertical, ZPTheme.spacing10)
                .background(.ultraThinMaterial, in: Capsule())
                .overlay(
                    Capsule()
                        .stroke(ZPTheme.panelBorderStrong, lineWidth: 1)
                )
                .shadow(color: ZPTheme.floatingShadow, radius: 12, y: 6)
                .padding(.bottom, ZPTheme.spacing20)
                .transition(.move(edge: .bottom).combined(with: .opacity))
                .onAppear {
                    dismissTask?.cancel()
                    dismissTask = Task {
                        try? await Task.sleep(nanoseconds: 1_500_000_000)
                        guard !Task.isCancelled else { return }
                        withAnimation(.easeOut(duration: 0.3)) {
                            isShowing = false
                        }
                    }
                }
            }
        }
        .animation(.spring(response: 0.3, dampingFraction: 0.8), value: isShowing)
    }
}

extension View {
    func toast(isShowing: Binding<Bool>, message: String, icon: String = "checkmark.circle.fill") -> some View {
        modifier(ToastOverlay(isShowing: isShowing, message: message, icon: icon))
    }
}
