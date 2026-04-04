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
        static let openVaultSheetTitle = "openVault.title"
        static let openVaultCancelButton = "openVault.cancelButton"
        static let bundleIdentifier = "com.tuanle.ZeroPass"
    }

    private var launchedApp: XCUIApplication?

    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    override func tearDownWithError() throws {
        if let launchedApp {
            terminate(launchedApp, failureContext: "after the UI test finished")
        }
        launchedApp = nil
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

    func testLaunchPerformance() throws {
        measure(metrics: [XCTApplicationLaunchMetric()]) {
            _ = launchFreshApp()
        }
    }

    private func launchFreshApp() -> XCUIApplication {
        ZeroPassUITestProcessPreflight.cleanStaleProcesses(
            failureMessage: "Failed to preflight stale ZeroPass processes before launching the UI test."
        )

        if let launchedApp {
            terminate(launchedApp, failureContext: "before relaunching the app in the same UI test")
            self.launchedApp = nil
        }

        let app = XCUIApplication()
        app.launchArguments += [
            "UITEST_MODE",
            "UITEST_RESET_STATE",
            "-ApplePersistenceIgnoreState",
            "YES"
        ]
        app.launch()
        app.activate()
        launchedApp = app
        return app
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
}
