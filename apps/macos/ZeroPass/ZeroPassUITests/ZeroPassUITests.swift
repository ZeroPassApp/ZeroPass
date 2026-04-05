//
//  ZeroPassUITests.swift
//  ZeroPassUITests
//
//  Created by Lê Anh Tuấn on 2/4/26.
//

import XCTest

enum ZeroPassUITestProcessPreflight {
    static func cleanStaleProcesses(
        failureMessage: String,
        file: StaticString = #filePath,
        line: UInt = #line
    ) {
        let cleanupCommands = [
            "/usr/bin/pkill -9 -f '/ZeroPass.app/Contents/MacOS/ZeroPass' >/dev/null 2>&1 || true",
            "/usr/bin/pkill -9 -f 'debugserver.*ZeroPass' >/dev/null 2>&1 || true"
        ]

        for command in cleanupCommands {
            _ = runShellCommand(command, file: file, line: line)
        }

        let deadline = Date().addingTimeInterval(5)
        while Date() < deadline {
            let survivors = staleProcessIdentifiers(file: file, line: line)
            if survivors.isEmpty {
                return
            }
            RunLoop.current.run(until: Date().addingTimeInterval(0.1))
        }

        let survivors = staleProcessIdentifiers(file: file, line: line)
        XCTAssertTrue(survivors.isEmpty, "\(failureMessage) Survivors: \(survivors)", file: file, line: line)
    }

    private static func staleProcessIdentifiers(
        file: StaticString,
        line: UInt
    ) -> [String] {
        let appPIDs = runShellCommand("/usr/bin/pgrep -f '/ZeroPass.app/Contents/MacOS/ZeroPass' || true", file: file, line: line)
        let debugserverPIDs = runShellCommand("/usr/bin/pgrep -f 'debugserver.*ZeroPass' || true", file: file, line: line)

        return [appPIDs, debugserverPIDs]
            .flatMap { output in
                output
                    .split(whereSeparator: \.isNewline)
                    .map(String.init)
            }
            .filter { !$0.isEmpty }
    }

    private static func runShellCommand(
        _ command: String,
        file: StaticString,
        line: UInt
    ) -> String {
        let process = Process()
        process.executableURL = URL(fileURLWithPath: "/bin/zsh")
        process.arguments = ["-lc", command]

        let outputPipe = Pipe()
        process.standardOutput = outputPipe
        process.standardError = Pipe()

        do {
            try process.run()
            process.waitUntilExit()
        } catch {
            XCTFail(
                "Failed to run shell command during UI test preflight: \(error.localizedDescription)",
                file: file,
                line: line
            )
            return ""
        }

        let outputData = outputPipe.fileHandleForReading.readDataToEndOfFile()
        return String(decoding: outputData, as: UTF8.self).trimmingCharacters(in: .whitespacesAndNewlines)
    }
}

final class ZeroPassUITests: XCTestCase {
    private enum UIElement {
        static let createVaultButton = "welcome.createVaultButton"
        static let openVaultButton = "welcome.openVaultButton"
        static let createVaultSheetTitle = "createVault.title"
        static let createVaultCancelButton = "createVault.cancelButton"
        static let createVaultChooseFolderButton = "createVault.chooseFolderButton"
        static let createVaultPasswordField = "createVault.masterPasswordField"
        static let createVaultConfirmPasswordField = "createVault.confirmPasswordField"
        static let createVaultSubmitButton = "createVault.submitButton"
        static let openVaultSheetTitle = "openVault.title"
        static let openVaultCancelButton = "openVault.cancelButton"
        static let openVaultChooseFolderButton = "openVault.chooseFolderButton"
        static let recoveryContinueButton = "recoveryPhrase.continueButton"
        static let recoveryConfirmSavedToggle = "recoveryPhrase.confirmSavedToggle"
        static let unlockPasswordField = "unlockVault.passwordField"
        static let unlockSubmitButton = "unlockVault.submitButton"
        static let mainShellRoot = "mainShell.root"
        static let mainShellNewItemButton = "mainShell.newItemButton"
        static let pickDirectoryEnvironmentKey = "UITEST_PICK_DIRECTORY_PATH"
    }

    private var launchedApp: XCUIApplication?
    private var temporaryDirectories: [URL] = []

    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    override func tearDownWithError() throws {
        if let launchedApp {
            terminate(launchedApp, failureContext: "after the UI test finished")
        }
        launchedApp = nil

        for directory in temporaryDirectories {
            try? FileManager.default.removeItem(at: directory)
        }
        temporaryDirectories.removeAll()
    }

    func testWelcomeScreenShowsPrimaryActionsOnFreshLaunch() throws {
        let app = launchFreshApp()

        XCTAssertTrue(app.buttons[UIElement.createVaultButton].waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons[UIElement.openVaultButton].waitForExistence(timeout: 5))
    }

    func testCreateVaultSheetCanBeOpenedAndCancelled() throws {
        let app = launchFreshApp()
        ensureMainWindowVisible(in: app)

        app.activate()
        focusMainWindow(in: app)
        app.typeKey(XCUIKeyboardKey.return, modifierFlags: [])

        let createSheetTitle = app.staticTexts[UIElement.createVaultSheetTitle]
        XCTAssertTrue(createSheetTitle.waitForExistence(timeout: 5))

        app.typeKey(XCUIKeyboardKey.escape, modifierFlags: [])

        XCTAssertTrue(waitForNonExistence(of: createSheetTitle, timeout: 5))
        XCTAssertTrue(app.buttons[UIElement.createVaultButton].waitForExistence(timeout: 5))
    }

    func testOpenVaultSheetCanBeOpenedAndCancelled() throws {
        let app = launchFreshApp()
        ensureMainWindowVisible(in: app)

        app.activate()
        focusMainWindow(in: app)
        app.typeKey("o", modifierFlags: .command)

        let openSheetTitle = app.staticTexts[UIElement.openVaultSheetTitle]
        XCTAssertTrue(openSheetTitle.waitForExistence(timeout: 5))

        app.typeKey(XCUIKeyboardKey.escape, modifierFlags: [])

        XCTAssertTrue(waitForNonExistence(of: openSheetTitle, timeout: 5))
        XCTAssertTrue(app.buttons[UIElement.openVaultButton].waitForExistence(timeout: 5))
    }

    func testOpenVaultCommandRevealsWindowAfterClosingWelcomeWindow() throws {
        let app = launchFreshApp()
        ensureMainWindowVisible(in: app)

        app.activate()
        focusMainWindow(in: app)
        app.typeKey("w", modifierFlags: .command)

        let createVaultButton = app.buttons[UIElement.createVaultButton]
        XCTAssertTrue(
            waitForNonExistence(of: createVaultButton, timeout: 5),
            "Expected the welcome window to close before exercising the recovery path."
        )

        app.activate()
        app.typeKey("o", modifierFlags: .command)

        let openSheetTitle = app.staticTexts[UIElement.openVaultSheetTitle]
        XCTAssertTrue(openSheetTitle.waitForExistence(timeout: 5))
        XCTAssertTrue(app.buttons[UIElement.openVaultCancelButton].waitForExistence(timeout: 5))

        app.typeKey(XCUIKeyboardKey.escape, modifierFlags: [])

        XCTAssertTrue(waitForNonExistence(of: openSheetTitle, timeout: 5))
        XCTAssertTrue(createVaultButton.waitForExistence(timeout: 5))
    }

    func testCreateVaultFlowShowsRecoveryPhraseAndContinuesToUnlockedShell() throws {
        let password = "ZeroPassUITestCreate1A"
        let vaultURL = makeTemporaryVaultURL(testName: #function)
        let app = launchFreshApp(resetState: true, pickedDirectory: vaultURL)

        openCreateVaultSheet(in: app)
        completeCreateVaultFlow(in: app, password: password)

        assertRecoveryPhrasePrimaryControlsVisible(
            in: app,
            screenDescription: "after creating a vault"
        )

        let continueButton = app.buttons[UIElement.recoveryContinueButton]
        let confirmSavedToggle = recoveryConfirmSavedToggle(in: app)
        confirmSavedToggle.click()

        continueButton.click()

        assertUnlockedShell(in: app)
        assertVaultFixtureExists(at: vaultURL)
    }

    func testRelaunchExistingVaultShowsLockedStateAndUnlocks() throws {
        let password = "ZeroPassUITestRelaunch2A"
        let vaultURL = makeTemporaryVaultURL(testName: #function)

        provisionVaultFixture(password: password, directory: vaultURL)

        let app = launchFreshApp(resetState: false)
        waitForLockedState(in: app)
        unlockVault(in: app, password: password)

        assertUnlockedShell(in: app)
    }

    func testOpenExistingVaultFlowUnlocksToMainShell() throws {
        let password = "ZeroPassUITestOpen3A"
        let vaultURL = makeTemporaryVaultURL(testName: #function)

        provisionVaultFixture(password: password, directory: vaultURL)

        let app = launchFreshApp(resetState: true, pickedDirectory: vaultURL)

        openOpenVaultSheet(in: app)

        let chooseFolderButton = app.buttons[UIElement.openVaultChooseFolderButton]
        XCTAssertTrue(
            chooseFolderButton.waitForExistence(timeout: 5),
            "Expected the open vault sheet to expose the deterministic folder picker action."
        )

        chooseFolderButton.click()

        waitForLockedState(in: app)
        unlockVault(in: app, password: password)

        assertUnlockedShell(in: app)
    }

    func testLaunchPerformance() throws {
        measure(metrics: [XCTApplicationLaunchMetric()]) {
            _ = launchFreshApp()
        }
    }

    private func launchFreshApp(
        resetState: Bool = true,
        pickedDirectory: URL? = nil
    ) -> XCUIApplication {
        ZeroPassUITestProcessPreflight.cleanStaleProcesses(
            failureMessage: "Failed to preflight stale ZeroPass processes before launching the UI test."
        )

        if let launchedApp {
            terminate(launchedApp, failureContext: "before relaunching the app in the same UI test")
            self.launchedApp = nil
        }

        let app = XCUIApplication()
        app.launchArguments += ["UITEST_MODE", "-ApplePersistenceIgnoreState", "YES"]
        if resetState {
            app.launchArguments.append("UITEST_RESET_STATE")
        }
        if let pickedDirectory {
            app.launchEnvironment[UIElement.pickDirectoryEnvironmentKey] = pickedDirectory.path
        }
        app.launch()
        app.activate()
        launchedApp = app
        return app
    }

    private func provisionVaultFixture(password: String, directory: URL) {
        let app = launchFreshApp(resetState: true, pickedDirectory: directory)

        openCreateVaultSheet(in: app)
        completeCreateVaultFlow(in: app, password: password)

        assertRecoveryPhrasePrimaryControlsVisible(
            in: app,
            screenDescription: "while provisioning a vault fixture"
        )

        let continueButton = app.buttons[UIElement.recoveryContinueButton]
        let confirmSavedToggle = recoveryConfirmSavedToggle(in: app)
        confirmSavedToggle.click()

        continueButton.click()

        assertUnlockedShell(in: app)
        assertVaultFixtureExists(at: directory)
    }

    private func openCreateVaultSheet(in app: XCUIApplication) {
        ensureMainWindowVisible(in: app)

        let button = app.buttons[UIElement.createVaultButton]
        XCTAssertTrue(button.waitForExistence(timeout: 5), "Expected the create vault action on the welcome screen.")

        button.click()

        XCTAssertTrue(
            app.staticTexts[UIElement.createVaultSheetTitle].waitForExistence(timeout: 5),
            "Expected the create vault sheet to appear."
        )
    }

    private func openOpenVaultSheet(in app: XCUIApplication) {
        ensureMainWindowVisible(in: app)

        let button = app.buttons[UIElement.openVaultButton]
        XCTAssertTrue(button.waitForExistence(timeout: 5), "Expected the open vault action on the welcome screen.")

        button.click()

        XCTAssertTrue(
            app.staticTexts[UIElement.openVaultSheetTitle].waitForExistence(timeout: 5),
            "Expected the open vault sheet to appear."
        )
    }

    private func completeCreateVaultFlow(in app: XCUIApplication, password: String) {
        let chooseFolderButton = app.buttons[UIElement.createVaultChooseFolderButton]
        XCTAssertTrue(
            chooseFolderButton.waitForExistence(timeout: 5),
            "Expected the create vault sheet to expose the folder picker action."
        )
        chooseFolderButton.click()

        let passwordField = app.secureTextFields[UIElement.createVaultPasswordField]
        XCTAssertTrue(passwordField.waitForExistence(timeout: 5), "Expected the master password field in the create vault sheet.")
        passwordField.click()
        passwordField.typeText(password)

        let confirmField = app.secureTextFields[UIElement.createVaultConfirmPasswordField]
        XCTAssertTrue(confirmField.waitForExistence(timeout: 5), "Expected the confirm password field in the create vault sheet.")
        confirmField.click()
        confirmField.typeText(password)

        let submitButton = app.buttons[UIElement.createVaultSubmitButton]
        XCTAssertTrue(submitButton.waitForExistence(timeout: 5), "Expected the create vault submit button.")
        XCTAssertTrue(submitButton.isEnabled, "Expected the create vault submit button to become enabled after entering matching credentials.")

        submitButton.click()
    }

    private func recoveryConfirmSavedToggle(in app: XCUIApplication) -> XCUIElement {
        let checkBox = app.checkBoxes[UIElement.recoveryConfirmSavedToggle]
        if checkBox.exists {
            return checkBox
        }

        let `switch` = app.switches[UIElement.recoveryConfirmSavedToggle]
        if `switch`.exists {
            return `switch`
        }

        return app.descendants(matching: .any)
            .matching(identifier: UIElement.recoveryConfirmSavedToggle)
            .firstMatch
    }

    private func assertRecoveryPhrasePrimaryControlsVisible(
        in app: XCUIApplication,
        screenDescription: String,
        timeout: TimeInterval = 15
    ) {
        let continueButton = app.buttons[UIElement.recoveryContinueButton]
        XCTAssertTrue(
            continueButton.waitForExistence(timeout: timeout),
            "Expected the recovery phrase screen to appear \(screenDescription)."
        )

        let confirmSavedToggle = recoveryConfirmSavedToggle(in: app)
        XCTAssertTrue(
            confirmSavedToggle.waitForExistence(timeout: 5),
            "Expected the recovery phrase confirmation toggle to appear \(screenDescription)."
        )

        XCTAssertTrue(
            waitForHittable(of: confirmSavedToggle, timeout: 5),
            "Expected the recovery phrase confirmation toggle to be visible without scrolling \(screenDescription)."
        )
        XCTAssertTrue(
            waitForHittable(of: continueButton, timeout: 5),
            "Expected the recovery phrase continue button to be visible without scrolling \(screenDescription)."
        )
    }

    private func waitForLockedState(in app: XCUIApplication) {
        let passwordField = waitForUnlockPasswordField(in: app, timeout: 15)
        XCTAssertTrue(passwordField.exists, "Expected the locked vault screen to show the password field.")
        XCTAssertTrue(
            app.buttons[UIElement.unlockSubmitButton].waitForExistence(timeout: 5),
            "Expected the locked vault screen to show the unlock action."
        )
    }

    private func unlockVault(in app: XCUIApplication, password: String) {
        let passwordField = waitForUnlockPasswordField(in: app, timeout: 15)
        passwordField.click()
        passwordField.typeText(password)

        let submitButton = app.buttons[UIElement.unlockSubmitButton]
        XCTAssertTrue(submitButton.waitForExistence(timeout: 5), "Expected the unlock action to be visible.")
        XCTAssertTrue(submitButton.isEnabled, "Expected the unlock button to be enabled after entering the password.")

        app.typeKey(XCUIKeyboardKey.escape, modifierFlags: [])
        app.typeKey(XCUIKeyboardKey.return, modifierFlags: [])
    }

    private func assertUnlockedShell(in app: XCUIApplication) {
        let newItemButton = app.buttons[UIElement.mainShellNewItemButton]
        XCTAssertTrue(
            newItemButton.waitForExistence(timeout: 15),
            "Expected the unlocked main shell to expose the New Item toolbar action."
        )

        XCTAssertTrue(
            app.otherElements[UIElement.mainShellRoot].exists || newItemButton.exists,
            "Expected the unlocked main shell marker to be present."
        )
    }

    private func assertVaultFixtureExists(at directory: URL) {
        let vaultMetadata = directory.appendingPathComponent("vault.json")
        XCTAssertTrue(
            FileManager.default.fileExists(atPath: vaultMetadata.path),
            "Expected the test vault fixture to exist at \(vaultMetadata.path)."
        )
    }

    private func makeTemporaryVaultURL(testName: String) -> URL {
        let sanitizedName = testName
            .replacingOccurrences(of: "[^A-Za-z0-9_-]+", with: "-", options: .regularExpression)
            .trimmingCharacters(in: CharacterSet(charactersIn: "-"))
        let directory = FileManager.default.temporaryDirectory
            .appendingPathComponent("ZeroPassUITests", isDirectory: true)
            .appendingPathComponent("\(sanitizedName)-\(UUID().uuidString)", isDirectory: true)

        temporaryDirectories.append(directory)
        return directory
    }

    private func waitForUnlockPasswordField(
        in app: XCUIApplication,
        timeout: TimeInterval
    ) -> XCUIElement {
        let secureField = app.secureTextFields[UIElement.unlockPasswordField]
        if secureField.waitForExistence(timeout: timeout) {
            return secureField
        }

        let plainField = app.textFields[UIElement.unlockPasswordField]
        XCTAssertTrue(
            plainField.waitForExistence(timeout: 2),
            "Expected the unlock password field to exist as either a secure or plain text field."
        )
        return plainField
    }

    private func terminate(_ app: XCUIApplication, failureContext: String) {
        guard app.state != .notRunning else {
            return
        }

        app.terminate()

        let deadline = Date().addingTimeInterval(5)
        while Date() < deadline {
            if app.state == .notRunning {
                return
            }
            RunLoop.current.run(until: Date().addingTimeInterval(0.1))
        }

        XCTAssertTrue(
            app.state == .notRunning,
            "Failed to terminate existing ZeroPass process \(failureContext). Final state: \(String(describing: app.state))"
        )
    }

    private func ensureMainWindowVisible(in app: XCUIApplication) {
        if app.buttons[UIElement.createVaultButton].waitForExistence(timeout: 5) {
            return
        }

        app.activate()
        app.typeKey("n", modifierFlags: .command)

        XCTAssertTrue(
            app.buttons[UIElement.createVaultButton].waitForExistence(timeout: 5),
            "Main window did not appear on launch and Command-N could not reveal one."
        )
    }

    private func focusMainWindow(in app: XCUIApplication) {
        let window = app.windows.firstMatch
        guard window.waitForExistence(timeout: 5) else {
            return
        }

        window.click()
    }

    private func waitForNonExistence(of element: XCUIElement, timeout: TimeInterval) -> Bool {
        let predicate = NSPredicate(format: "exists == false")
        let expectation = XCTNSPredicateExpectation(predicate: predicate, object: element)
        return XCTWaiter().wait(for: [expectation], timeout: timeout) == .completed
    }

    private func waitForHittable(of element: XCUIElement, timeout: TimeInterval) -> Bool {
        let predicate = NSPredicate(format: "hittable == true")
        let expectation = XCTNSPredicateExpectation(predicate: predicate, object: element)
        return XCTWaiter().wait(for: [expectation], timeout: timeout) == .completed
    }
}
