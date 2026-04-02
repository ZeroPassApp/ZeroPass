//
//  ContentView.swift
//  ZeroPass
//
//  Created by Lê Anh Tuấn on 2/4/26.
//

import SwiftUI

struct ContentView: View {
    @EnvironmentObject var vault: VaultClient

    var body: some View {
        Group {
            switch vault.state {
            case .noVault:
                WelcomeView()
            case .locked:
                UnlockVaultView()
            case .showingRecovery(let mnemonic):
                RecoveryPhraseView(mnemonic: mnemonic)
            case .unlocked:
                MainShellView()
            }
        }
        .task {
            await vault.restoreLastVaultIfAvailable()
        }
    }
}

#Preview {
    ContentView()
        .environmentObject(VaultClient())
}
