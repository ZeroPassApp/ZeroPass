import { useEffect } from "react";
import { useAuthStore } from "./stores/auth-store";
import { WelcomePage } from "./pages/welcome-page";
import { UnlockPage } from "./pages/unlock-page";
import { RecoveryPhrasePage } from "./pages/recovery-phrase-page";
import { VaultPage } from "./pages/vault-page";
import { ToastOverlay } from "./components/common/toast-overlay";

function App() {
    const authState = useAuthStore((s) => s.state);

    // On mount, restore last opened vault or show welcome
    useEffect(() => {
        useAuthStore.getState().restore();
    }, []);

    return (
        <>
            {authState === "restoring" && <LoadingScreen />}
            {authState === "noVault" && <WelcomePage />}
            {authState === "locked" && <UnlockPage />}
            {authState === "showingRecovery" && <RecoveryPhrasePage />}
            {authState === "unlocked" && <VaultPage />}
            <ToastOverlay />
        </>
    );
}

function LoadingScreen() {
    return (
        <div
            style={{
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                height: "100%",
                background: "var(--color-bg)",
            }}
        >
            <span style={{ color: "var(--color-text-tertiary)" }}>Loading…</span>
        </div>
    );
}

export default App;
