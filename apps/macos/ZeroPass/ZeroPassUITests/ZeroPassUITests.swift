//
//  ZeroPassUITests.swift
//  ZeroPassUITests
//
//  Created by Lê Anh Tuấn on 2/4/26.
//

import XCTest
import AppKit

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

    override func setUpWithError() throws {
        continueAfterFailure = false
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
        app.typeKey("o", modifierFlags: .command)

        let openSheetTitle = app.staticTexts[UIElement.openVaultSheetTitle]
        XCTAssertTrue(openSheetTitle.waitForExistence(timeout: 5))

        app.typeKey(XCUIKeyboardKey.escape, modifierFlags: [])

        XCTAssertTrue(waitForNonExistence(of: openSheetTitle, timeout: 5))
        XCTAssertTrue(app.buttons[UIElement.openVaultButton].waitForExistence(timeout: 5))
    }

    func testLaunchPerformance() throws {
        measure(metrics: [XCTApplicationLaunchMetric()]) {
            launchFreshApp()
        }
    }

    private func launchFreshApp() -> XCUIApplication {
        terminateRunningAppIfNeeded()

        let app = XCUIApplication()
        app.launchArguments += [
            "UITEST_MODE",
            "UITEST_RESET_STATE",
            "-ApplePersistenceIgnoreState",
            "YES"
        ]
        app.launch()
        app.activate()
        return app
    }

    private func terminateRunningAppIfNeeded() {
        let runningApplications = NSRunningApplication.runningApplications(withBundleIdentifier: UIElement.bundleIdentifier)

        for runningApplication in runningApplications {
            _ = runningApplication.forceTerminate()
        }

        let deadline = Date().addingTimeInterval(5)
        while Date() < deadline {
            let stillRunning = NSRunningApplication.runningApplications(withBundleIdentifier: UIElement.bundleIdentifier)
            if stillRunning.isEmpty {
                return
            }
            RunLoop.current.run(until: Date().addingTimeInterval(0.1))
        }
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

    private func waitForNonExistence(of element: XCUIElement, timeout: TimeInterval) -> Bool {
        let predicate = NSPredicate(format: "exists == false")
        let expectation = XCTNSPredicateExpectation(predicate: predicate, object: element)
        return XCTWaiter().wait(for: [expectation], timeout: timeout) == .completed
    }
}
