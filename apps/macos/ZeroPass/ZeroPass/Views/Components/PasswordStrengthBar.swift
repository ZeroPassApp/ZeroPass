import SwiftUI

struct PasswordStrengthBar: View {
    let score: Int

    private var filledSegments: Int { min(max(score + 1, 0), 4) }

    private func segmentColor(for index: Int) -> Color {
        guard index < filledSegments else { return ZPTheme.separator }
        switch score {
        case 0: return ZPTheme.destructive
        case 1: return ZPTheme.warning
        case 2: return ZPTheme.accent
        case 3: return ZPTheme.success
        default: return ZPTheme.success
        }
    }

    var body: some View {
        HStack(spacing: ZPTheme.spacing6) {
            ForEach(0..<4, id: \.self) { index in
                RoundedRectangle(cornerRadius: 2)
                    .fill(segmentColor(for: index))
                    .frame(height: 6)
            }
        }
        .animation(.easeInOut(duration: 0.25), value: score)
    }
}
