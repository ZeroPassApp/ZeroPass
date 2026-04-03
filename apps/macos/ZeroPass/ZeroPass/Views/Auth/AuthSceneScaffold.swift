import SwiftUI

enum AuthMessageTone {
    case info
    case success
    case warning
    case error

    var foregroundColor: Color {
        switch self {
        case .info:
            return ZPTheme.info
        case .success:
            return ZPTheme.success
        case .warning:
            return ZPTheme.warning
        case .error:
            return ZPTheme.destructive
        }
    }

    var backgroundColor: Color {
        switch self {
        case .info:
            return ZPTheme.infoFill
        case .success:
            return ZPTheme.successFill
        case .warning:
            return ZPTheme.warningFill
        case .error:
            return ZPTheme.errorFill
        }
    }
}

struct AuthMessageView: View {
    let text: String
    let systemImage: String
    var tone: AuthMessageTone = .info

    var body: some View {
        HStack(alignment: .top, spacing: ZPTheme.spacing8) {
            Image(systemName: systemImage)
                .font(.subheadline.weight(.semibold))
                .foregroundStyle(tone.foregroundColor)
                .padding(.top, 1)

            Text(text)
                .font(.callout)
                .foregroundStyle(ZPTheme.textPrimary)
                .fixedSize(horizontal: false, vertical: true)
        }
        .padding(.horizontal, ZPTheme.spacing12)
        .padding(.vertical, ZPTheme.spacing10)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(tone.backgroundColor, in: RoundedRectangle(cornerRadius: ZPTheme.radiusMedium, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: ZPTheme.radiusMedium, style: .continuous)
                .stroke(tone.foregroundColor.opacity(0.18), lineWidth: 1)
        )
    }
}

struct AuthSupportingNoteView: View {
    let text: String
    let systemImage: String

    var body: some View {
        Label {
            Text(text)
                .font(.footnote)
                .foregroundStyle(ZPTheme.textSecondary)
                .fixedSize(horizontal: false, vertical: true)
        } icon: {
            Image(systemName: systemImage)
                .foregroundStyle(ZPTheme.textSecondary)
        }
    }
}

struct AuthSceneScaffold<Accessory: View, Content: View, Footer: View>: View {
    private let title: String
    private let subtitle: String?
    private let detail: String?
    private let detailSymbolName: String?
    private let symbolName: String
    private let symbolTint: Color
    private let accessory: Accessory
    private let content: Content
    private let footer: Footer

    init(
        title: String,
        subtitle: String? = nil,
        detail: String? = nil,
        detailSymbolName: String? = nil,
        symbolName: String,
        symbolTint: Color = .accentColor,
        @ViewBuilder accessory: () -> Accessory,
        @ViewBuilder content: () -> Content,
        @ViewBuilder footer: () -> Footer
    ) {
        self.title = title
        self.subtitle = subtitle
        self.detail = detail
        self.detailSymbolName = detailSymbolName
        self.symbolName = symbolName
        self.symbolTint = symbolTint
        self.accessory = accessory()
        self.content = content()
        self.footer = footer()
    }

    var body: some View {
        ZStack {
            backgroundLayer

            VStack {
                VStack(alignment: .leading, spacing: ZPTheme.spacing24) {
                    header
                    content
                    footer
                }
                .frame(maxWidth: ZPTheme.authPanelMaxWidth, alignment: .leading)
                .padding(ZPTheme.spacing32)
                .background(
                    RoundedRectangle(cornerRadius: ZPTheme.radiusXLarge, style: .continuous)
                        .fill(ZPTheme.authPanelBackground)
                )
                .overlay(
                    RoundedRectangle(cornerRadius: ZPTheme.radiusXLarge, style: .continuous)
                        .stroke(ZPTheme.authPanelBorder, lineWidth: 1)
                )
                .shadow(color: ZPTheme.panelShadow, radius: 20, y: 12)
            }
            .padding(ZPTheme.spacing24)
            .frame(maxWidth: .infinity, maxHeight: .infinity)
        }
    }

    private var backgroundLayer: some View {
        ZPTheme.authSceneBackground
            .ignoresSafeArea()
    }

    private var header: some View {
        HStack(alignment: .top, spacing: ZPTheme.spacing16) {
            ZStack {
                RoundedRectangle(cornerRadius: ZPTheme.radiusMedium, style: .continuous)
                    .fill(symbolTint.opacity(0.14))
                    .frame(width: 44, height: 44)

                Image(systemName: symbolName)
                    .font(.title3.weight(.semibold))
                    .foregroundStyle(symbolTint)
                    .accessibilityHidden(true)
            }

            VStack(alignment: .leading, spacing: ZPTheme.spacing4) {
                Text(title)
                    .font(.title2)
                    .fontWeight(.semibold)
                    .foregroundStyle(ZPTheme.textPrimary)

                if let subtitle, !subtitle.isEmpty {
                    Text(subtitle)
                        .font(.subheadline)
                        .fontWeight(.semibold)
                        .foregroundStyle(ZPTheme.textPrimary)
                }

                if let detail, !detail.isEmpty {
                    if let detailSymbolName, !detailSymbolName.isEmpty {
                        Label(detail, systemImage: detailSymbolName)
                            .font(.footnote)
                            .foregroundStyle(ZPTheme.textSecondary)
                            .lineLimit(1)
                            .truncationMode(.middle)
                            .labelStyle(.titleAndIcon)
                    } else {
                        Text(detail)
                            .font(.footnote)
                            .foregroundStyle(ZPTheme.textSecondary)
                            .lineLimit(1)
                            .truncationMode(.middle)
                    }
                }
            }

            Spacer(minLength: ZPTheme.spacing16)
            accessory
        }
    }
}

extension AuthSceneScaffold where Accessory == EmptyView, Footer == EmptyView {
    init(
        title: String,
        subtitle: String? = nil,
        detail: String? = nil,
        detailSymbolName: String? = nil,
        symbolName: String,
        symbolTint: Color = .accentColor,
        @ViewBuilder content: () -> Content
    ) {
        self.init(
            title: title,
            subtitle: subtitle,
            detail: detail,
            detailSymbolName: detailSymbolName,
            symbolName: symbolName,
            symbolTint: symbolTint,
            accessory: { EmptyView() },
            content: content,
            footer: { EmptyView() }
        )
    }
}

extension AuthSceneScaffold where Footer == EmptyView {
    init(
        title: String,
        subtitle: String? = nil,
        detail: String? = nil,
        detailSymbolName: String? = nil,
        symbolName: String,
        symbolTint: Color = .accentColor,
        @ViewBuilder accessory: () -> Accessory,
        @ViewBuilder content: () -> Content
    ) {
        self.init(
            title: title,
            subtitle: subtitle,
            detail: detail,
            detailSymbolName: detailSymbolName,
            symbolName: symbolName,
            symbolTint: symbolTint,
            accessory: accessory,
            content: content,
            footer: { EmptyView() }
        )
    }
}

extension AuthSceneScaffold where Accessory == EmptyView {
    init(
        title: String,
        subtitle: String? = nil,
        detail: String? = nil,
        detailSymbolName: String? = nil,
        symbolName: String,
        symbolTint: Color = .accentColor,
        @ViewBuilder content: () -> Content,
        @ViewBuilder footer: () -> Footer
    ) {
        self.init(
            title: title,
            subtitle: subtitle,
            detail: detail,
            detailSymbolName: detailSymbolName,
            symbolName: symbolName,
            symbolTint: symbolTint,
            accessory: { EmptyView() },
            content: content,
            footer: footer
        )
    }
}
