# Cucumber-js

This is the last version in the `cucumber` series. New releases will henceforth
be made under the `@cucumber/cucumber` package name.

By installing this version you are implicitly installing the new `@cucumber/cucumber`
package.

You should `npm uninstall cucumber` and then `npm install --save-dev @cucumber/cucumber`.

After doing this you should replace your `require` or `import` statements to load
`@cucumber/cucumber` instead of `cucumber`.

If you are using TypeScript, `npm uninstall @types/cucumber` (v7 has types built-in). Also
replace `TableDefinition` with `DataTable` if you are using these in your code.
