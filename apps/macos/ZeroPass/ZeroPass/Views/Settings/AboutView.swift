import SwiftUI

struct AboutView: View {
    private var versionString: String {
        let v = Bundle.main.object(forInfoDictionaryKey: "CFBundleShortVersionString") as? String ?? ""
        let b = Bundle.main.object(forInfoDictionaryKey: "CFBundleVersion") as? String ?? ""
        if b.isEmpty { return v }
        if v.isEmpty { return b }
        return "\(v) (\(b))"
    }

    var body: some View {
        VStack(spacing: 16) {
            Spacer()

            ZStack {
                RoundedRectangle(cornerRadius: ZPTheme.radiusXLarge, style: .continuous)
                    .fill(ZPTheme.accentSoft)

                Image(systemName: "lock.shield.fill")
                    .font(.system(size: 40, weight: .semibold))
                    .foregroundStyle(ZPTheme.accent)
            }
            .frame(width: 88, height: 88)

            VStack(spacing: 4) {
                Text("ZeroPass")
                    .font(.title)
                    .bold()
                    .foregroundStyle(ZPTheme.textPrimary)

                Text("Version \(versionString)")
                    .font(.callout)
                    .foregroundStyle(ZPTheme.textSecondary)
            }

            Text("A zero-knowledge, local-first credential manager.\nPowered by a Go core with a native SwiftUI interface.")
                .font(.callout)
                .foregroundStyle(ZPTheme.textSecondary)
                .multilineTextAlignment(.center)
                .fixedSize(horizontal: false, vertical: true)

            Divider()
                .frame(maxWidth: 200)

            VStack(spacing: 4) {
                Link("GitHub Repository", destination: URL(string: "https://github.com/nicholasruunu/ZeroPass")!)
                    .font(.callout)

                Text("macOS \(ProcessInfo.processInfo.operatingSystemVersionString)")
                    .font(.caption)
                    .foregroundStyle(ZPTheme.textTertiary)
            }

            Spacer()
        }
        .frame(maxWidth: .infinity)
        .padding(20)
        .background(ZPTheme.workspaceBackground)
    }
}
