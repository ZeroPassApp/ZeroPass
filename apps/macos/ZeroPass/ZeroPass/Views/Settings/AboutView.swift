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

            Image(systemName: "lock.shield.fill")
                .font(.system(size: 48))
                .foregroundStyle(.tint)

            VStack(spacing: 4) {
                Text("ZeroPass")
                    .font(.title)
                    .bold()

                Text("Version \(versionString)")
                    .font(.callout)
                    .foregroundStyle(.secondary)
            }

            Text("A zero-knowledge, local-first credential manager.\nPowered by a Go core with a native SwiftUI interface.")
                .font(.callout)
                .foregroundStyle(.secondary)
                .multilineTextAlignment(.center)
                .fixedSize(horizontal: false, vertical: true)

            Divider()
                .frame(maxWidth: 200)

            VStack(spacing: 4) {
                Link("GitHub Repository", destination: URL(string: "https://github.com/nicholasruunu/ZeroPass")!)
                    .font(.callout)

                Text("macOS \(ProcessInfo.processInfo.operatingSystemVersionString)")
                    .font(.caption)
                    .foregroundStyle(.tertiary)
            }

            Spacer()
        }
        .frame(maxWidth: .infinity)
        .padding(20)
    }
}
