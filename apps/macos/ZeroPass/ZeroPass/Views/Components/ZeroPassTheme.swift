import AppKit
import SwiftUI

enum ZPSurfaceStyle {
    case card
    case elevated
    case inset
    case floating
    case muted
    case selected
    case accent
}

// MARK: - ZeroPass Design Tokens

enum ZPTheme {

    // MARK: Product Colors

    static let accent = dynamicColor(light: "#0A84FF", dark: "#4EA1FF")
    static let accentSoft = accent.opacity(0.16)
    static let accentStrong = accent.opacity(0.24)
    static let accentGlow = accent.opacity(0.18)
    static let unlockButton = dynamicColor(light: "#1A73E8", dark: "#3B8AF6")

    static let authSceneBackground = dynamicColor(light: "#F4F7FA", dark: "#121518")
    static let authSceneAccent = accent.opacity(0.08)
    static let workspaceBackground = dynamicColor(light: "#EEF2F5", dark: "#14181B")

    static let panelBackground = dynamicColor(light: "#FFFFFF", dark: "#1C2126")
    static let panelBackgroundElevated = dynamicColor(light: "#FFFFFF", dark: "#22282E")
    static let panelBackgroundMuted = dynamicColor(light: "#F6F8FB", dark: "#191D22")
    static let inputBackground = dynamicColor(light: "#F2F5F8", dark: "#15191D")
    static let chipBackground = dynamicColor(light: "#F1F4F8", dark: "#20262C")
    static let sidebarBackground = dynamicColor(light: "#F6F8FA", dark: "#171B1F")

    static let panelBorder = dynamicColor(light: "#D9E0E8", dark: "#2C343B")
    static let panelBorderStrong = dynamicColor(light: "#CCD5DE", dark: "#39434C")
    static let separator = dynamicColor(light: "#E4EAF0", dark: "#262D33")
    static let separatorSubtle = dynamicColor(light: "#EDF1F5", dark: "#20262B")

    static let selectionFill = accent.opacity(0.18)
    static let selectionStroke = accent.opacity(0.38)
    static let pillBackground = dynamicColor(light: "#EDF4FF", dark: "#18212C")
    static let pillBorder = accent.opacity(0.28)

    static let textPrimary = dynamicColor(light: "#111418", dark: "#F3F7FB")
    static let textSecondary = dynamicColor(light: "#556371", dark: "#AAB5C0")
    static let textTertiary = dynamicColor(light: "#748291", dark: "#808D98")
    static let textMuted = dynamicColor(light: "#8996A3", dark: "#66727D")

    static let info = accent
    static let success = dynamicColor(light: "#0F9D58", dark: "#32D17D")
    static let warning = dynamicColor(light: "#B97800", dark: "#F4B544")
    static let destructive = dynamicColor(light: "#D93025", dark: "#FF6B60")

    static let infoFill = info.opacity(0.12)
    static let successFill = success.opacity(0.12)
    static let warningFill = warning.opacity(0.14)
    static let errorFill = destructive.opacity(0.12)

    static let panelShadow = Color.black.opacity(0.18)
    static let floatingShadow = Color.black.opacity(0.28)
    static let fieldShadow = Color.black.opacity(0.04)

    static let authPanelBackground = panelBackgroundElevated
    static let authPanelBorder = panelBorder
    static let authInsetBackground = inputBackground
    static let authInsetBorder = panelBorderStrong

    // MARK: Spacing

    static let spacing4: CGFloat = 4
    static let spacing6: CGFloat = 6
    static let spacing8: CGFloat = 8
    static let spacing10: CGFloat = 10
    static let spacing12: CGFloat = 12
    static let spacing14: CGFloat = 14
    static let spacing16: CGFloat = 16
    static let spacing18: CGFloat = 18
    static let spacing20: CGFloat = 20
    static let spacing24: CGFloat = 24
    static let spacing28: CGFloat = 28
    static let spacing32: CGFloat = 32
    static let spacing40: CGFloat = 40

    // MARK: Radii

    static let radiusSmall: CGFloat = 10
    static let radiusMedium: CGFloat = 12
    static let radiusLarge: CGFloat = 16
    static let radiusXLarge: CGFloat = 22

    // MARK: Sizing

    static let authPanelMaxWidth: CGFloat = 620
    static let authFieldHeight: CGFloat = 42
    static let authEditorMinHeight: CGFloat = 132
    static let authSheetWidth: CGFloat = 560

    static func dynamicColor(light: String, dark: String) -> Color {
        Color(nsColor: NSColor(name: nil) { appearance in
            let useDark = appearance.bestMatch(from: [.darkAqua, .aqua]) == .darkAqua
            return NSColor(hex: useDark ? dark : light) ?? (useDark ? .windowBackgroundColor : .white)
        })
    }
}

extension View {
    func zpSurface(_ style: ZPSurfaceStyle = .card, radius: CGFloat = ZPTheme.radiusLarge, shadow: Bool = true) -> some View {
        modifier(ZPSurfaceModifier(style: style, radius: radius, shadow: shadow))
    }
}

private struct ZPSurfaceModifier: ViewModifier {
    let style: ZPSurfaceStyle
    let radius: CGFloat
    let shadow: Bool

    func body(content: Content) -> some View {
        content
            .background(backgroundColor, in: RoundedRectangle(cornerRadius: radius, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: radius, style: .continuous)
                    .strokeBorder(strokeColor, lineWidth: strokeWidth)
            )
            .shadow(color: shadow ? shadowColor : .clear, radius: shadowRadius, y: shadowOffset)
    }

    private var backgroundColor: Color {
        switch style {
        case .card:
            return ZPTheme.panelBackground
        case .elevated:
            return ZPTheme.panelBackgroundElevated
        case .inset:
            return ZPTheme.inputBackground
        case .floating:
            return ZPTheme.panelBackgroundElevated.opacity(0.94)
        case .muted:
            return ZPTheme.panelBackgroundMuted
        case .selected:
            return ZPTheme.selectionFill
        case .accent:
            return ZPTheme.accentSoft
        }
    }

    private var strokeColor: Color {
        switch style {
        case .selected:
            return ZPTheme.selectionStroke
        case .accent:
            return ZPTheme.pillBorder
        case .elevated, .floating:
            return ZPTheme.panelBorderStrong
        case .card, .muted:
            return ZPTheme.panelBorder
        case .inset:
            return ZPTheme.authInsetBorder
        }
    }

    private var strokeWidth: CGFloat {
        style == .selected ? 1.2 : 1
    }

    private var shadowColor: Color {
        switch style {
        case .floating:
            return ZPTheme.floatingShadow
        case .elevated, .card:
            return ZPTheme.panelShadow
        default:
            return .clear
        }
    }

    private var shadowRadius: CGFloat {
        switch style {
        case .floating:
            return 24
        case .elevated:
            return 16
        case .card:
            return 10
        default:
            return 0
        }
    }

    private var shadowOffset: CGFloat {
        style == .floating ? 12 : 6
    }
}

private extension NSColor {
    convenience init?(hex: String) {
        var sanitized = hex.trimmingCharacters(in: .whitespacesAndNewlines)
        sanitized = sanitized.replacingOccurrences(of: "#", with: "")

        guard sanitized.count == 6 || sanitized.count == 8 else {
            return nil
        }

        var value: UInt64 = 0
        guard Scanner(string: sanitized).scanHexInt64(&value) else {
            return nil
        }

        let red: CGFloat
        let green: CGFloat
        let blue: CGFloat
        let alpha: CGFloat

        if sanitized.count == 8 {
            red = CGFloat((value & 0xFF00_0000) >> 24) / 255
            green = CGFloat((value & 0x00FF_0000) >> 16) / 255
            blue = CGFloat((value & 0x0000_FF00) >> 8) / 255
            alpha = CGFloat(value & 0x0000_00FF) / 255
        } else {
            red = CGFloat((value & 0xFF00_00) >> 16) / 255
            green = CGFloat((value & 0x00FF_00) >> 8) / 255
            blue = CGFloat(value & 0x0000_FF) / 255
            alpha = 1
        }

        self.init(red: red, green: green, blue: blue, alpha: alpha)
    }
}
