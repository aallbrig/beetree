import js from '@eslint/js';
import tseslint from 'typescript-eslint';

export default tseslint.config(
  js.configs.recommended,
  ...tseslint.configs.strictTypeChecked,
  {
    languageOptions: {
      parserOptions: { projectService: true, tsconfigRootDir: import.meta.dirname },
    },
    rules: {
      // the escape hatch stays closed: describe the shape instead
      '@typescript-eslint/no-explicit-any': 'error',
      '@typescript-eslint/no-unsafe-assignment': 'error',
      '@typescript-eslint/no-unsafe-member-access': 'error',
      '@typescript-eslint/no-unsafe-argument': 'error',
      '@typescript-eslint/no-unsafe-call': 'error',
      '@typescript-eslint/no-unsafe-return': 'error',
      '@typescript-eslint/no-non-null-assertion': 'error',
      '@typescript-eslint/restrict-template-expressions': ['error', { allowNumber: true }],
      'no-restricted-syntax': [
        'error',
        {
          selector: "TSAsExpression > TSUnknownKeyword",
          message: 'Do not launder types through `as unknown as`; narrow with a type guard.',
        },
        {
          selector: "AssignmentExpression[left.property.name='innerHTML']",
          message: 'Build DOM with h()/textContent; never assign innerHTML (XSS).',
        },
      ],
    },
  },
  { ignores: ['dist', 'node_modules', '*.config.js'] },
);
