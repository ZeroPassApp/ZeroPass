import SwiftUI

struct SettingsView: View {
    enum Tab: Hashable {
        case general
        case security
        case sync
        case about
    }

    @State private var tab: Tab = .general

    var body: some View {
        TabView(selection: $tab) {
            GeneralSettingsView()
                .tabItem { Label("General", systemImage: "gear") }
                .tag(Tab.general)

            SecuritySettingsView()
                .tabItem { Label("Security", systemImage: "lock") }
                .tag(Tab.security)

            SyncSettingsView()
                .tabItem { Label("Sync", systemImage: "arrow.triangle.2.circlepath") }
                .tag(Tab.sync)

            AboutView()
                .tabItem { Label("About", systemImage: "info.circle") }
                .tag(Tab.about)
        }
        .padding(20)
        .frame(minWidth: 500, idealWidth: 550, minHeight: 400, idealHeight: 480)
    }
}
