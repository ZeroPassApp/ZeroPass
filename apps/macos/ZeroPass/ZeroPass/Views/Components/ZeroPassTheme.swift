import AppKit
import SwiftUI

// MARK: - ZeroPass Design Tokens

enum ZPTheme {

    // MARK: Colors

    static let authSceneBackground = Color.clear
    static let authSceneAccent = Color.accentColor.opacity(0.03)
    static let authPanelBackground = Color(nsColor: .controlBackgroundColor)
    static let authPanelBorder = Color(nsColor: .separatorColor).opacity(0.42)
    static let authInsetBackground = Color(nsColor: .textBackgroundColor)
    static let authInsetBorder = Color(nsColor: .separatorColor)

    static let textPrimary = Color.primary
    static let textSecondary = Color.secondary
    static let textTertiary = Color.secondary.opacity(0.85)
    static let textMuted = Color.secondary

    static let info = Color.accentColor
    static let success = Color(nsColor: .systemGreen)
    static let warning = Color(nsColor: .systemOrange)
    static let destructive = Color(nsColor: .systemRed)

    static let infoFill = info.opacity(0.10)
    static let successFill = success.opacity(0.10)
    static let warningFill = warning.opacity(0.12)
    static let errorFill = destructive.opacity(0.10)

    static let panelShadow = Color.black.opacity(0.16)
    static let fieldShadow = Color.black.opacity(0.03)

    // MARK: Spacing

    static let spacing4: CGFloat = 4
    static let spacing6: CGFloat = 6
    static let spacing8: CGFloat = 8
    static let spacing10: CGFloat = 10
    static let spacing12: CGFloat = 12
    static let spacing16: CGFloat = 16
    static let spacing20: CGFloat = 20
    static let spacing24: CGFloat = 24
    static let spacing32: CGFloat = 32

    // MARK: Radii

    static let radiusSmall: CGFloat = 8
    static let radiusMedium: CGFloat = 12
    static let radiusLarge: CGFloat = 16
    static let radiusXLarge: CGFloat = 24

    // MARK: Sizing

    static let authPanelMaxWidth: CGFloat = 560
    static let authFieldHeight: CGFloat = 40
    static let authEditorMinHeight: CGFloat = 120
    static let authSheetWidth: CGFloat = 460
}
