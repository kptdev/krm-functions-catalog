module.exports = {
  root: true,
  parser: '@typescript-eslint/parser',
  parserOptions: {
    ecmaVersion: 2020,
    sourceType: 'module',
    project: './tsconfig.json',
    tsconfigRootDir: __dirname,
  },
  plugins: [
    '@typescript-eslint',
    'import',
    'jsdoc',
    'prefer-arrow',
    'no-null',
  ],
  extends: [
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
    'plugin:import/recommended',
    'plugin:import/typescript',
    'plugin:jsdoc/recommended',
    'prettier',
  ],
  env: {
    es6: true,
    node: true,
    jasmine: true,
  },
  ignorePatterns: ['**/gen/**', '**/*.json', 'node_modules/**'],
  rules: {
    // Prefer-arrow rule
    'prefer-arrow/prefer-arrow-functions': 'off',
    '@typescript-eslint/array-type': ['error', { default: 'array-simple' }],
    '@typescript-eslint/ban-ts-comment': ['error', { 'ts-ignore': true }],
    '@typescript-eslint/ban-types': [
      'error',
      {
        extendDefaults: true,
        types: {
          Object: { message: 'Use {} instead.' },
          String: { message: "Use 'string' instead." },
          Number: { message: "Use 'number' instead." },
          Boolean: { message: "Use 'boolean' instead." },
        },
      },
    ],
    '@typescript-eslint/explicit-member-accessibility': ['error', { accessibility: 'no-public' }],
    "@typescript-eslint/naming-convention": [
            "error",
            {
                "selector": "variable",
                "format": [
                    "camelCase",
                    "UPPER_CASE"
                ],
                "leadingUnderscore": "allow",
                "trailingUnderscore": "allow"
            }
        ],
    '@typescript-eslint/no-explicit-any': 'error',
    '@typescript-eslint/no-inferrable-types': 'error',
    'jsdoc/require-param-description': 'off',
    'jsdoc/require-param-type': 'off',
    'jsdoc/require-returns': 'off',
    'curly': ['error', 'multi-line'],
    'eqeqeq': ['error', 'always', { null: 'ignore' }],
    'no-console': 'off',
    'no-debugger': 'error',
    'no-var': 'error',
    'prefer-const': 'error',
    'semi': ['error', 'always'],
    'comma-dangle': [
      'error',
      {
        arrays: 'always-multiline',
        objects: 'always-multiline',
        functions: 'never',
      },
    ],
    'use-isnan': 'error',
    'no-restricted-syntax': [
      'error',
      {
        selector: "CallExpression[callee.name='parseInt']",
        message: 'tsstyle#type-coercion',
      },
      {
        selector: "CallExpression[callee.name='parseFloat']",
        message: 'tsstyle#type-coercion',
      },
      {
        selector: "NewExpression[callee.name='Array']",
        message: 'tsstyle#array-constructor',
      },
      {
        selector: "[property.name='innerText']",
        message: 'Use .textContent instead. tsstyle#browser-oddities',
      },
    ],
    'no-null/no-null': 'error'
  },
  settings: {
    react: {
      version: 'detect',
    },
  },
};
