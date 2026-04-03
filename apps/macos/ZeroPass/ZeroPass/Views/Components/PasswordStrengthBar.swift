import SwiftUI

struct PasswordStrengthBar: View {
    let score: Int

    private var filledSegments: Int { min(max(score + 1, 0), 4) }

    private func segmentColor(for index: Int) -> Color {
        guard index < filledSegments else { return .gray.opacity(0.2) }
        switch score {
        case 0: return .red
        case 1: return .orange
        case 2: return .yellow
        case 3: return .green
        default: return .teal
        }
    }

    var body: some View {
        HStack(spacing: 4) {
            ForEach(0..<4, id: \.self) { index in
                RoundedRectangle(cornerRadius: 2)
                    .fill(segmentColor(for: index))
                    .frame(height: 4)
            }
        }
        .animation(.easeInOut(duration: 0.25), value: score)
    }
}
