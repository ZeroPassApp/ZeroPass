import SwiftUI

struct ToastOverlay: ViewModifier {
    @Binding var isShowing: Bool
    let message: String
    let icon: String
    @State private var dismissTask: Task<Void, Never>?

    func body(content: Content) -> some View {
        content.overlay(alignment: .bottom) {
            if isShowing {
                HStack(spacing: 6) {
                    Image(systemName: icon)
                        .font(.system(size: 12, weight: .medium))
                    Text(message)
                        .font(.system(size: 12, weight: .medium))
                }
                .padding(.horizontal, 14)
                .padding(.vertical, 8)
                .background(.regularMaterial)
                .clipShape(Capsule())
                .shadow(color: .black.opacity(0.1), radius: 4, y: 2)
                .padding(.bottom, 16)
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
