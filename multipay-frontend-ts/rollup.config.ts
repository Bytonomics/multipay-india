import typescript from '@rollup/plugin-typescript';
import resolve from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import terser from '@rollup/plugin-terser';
import postcss from 'rollup-plugin-postcss';

export default [
  // Core entry (no React dependency)
  {
    input: 'src/core/index.ts',
    output: {
      dir: 'dist/core',
      format: 'es',
      exports: 'named',
      preserveModules: true,
      preserveModulesRoot: 'src/core',
      entryFileNames: '[name].mjs',
      sourcemap: true,
    },
    external: [], // No peer dependencies for core
    plugins: [
      resolve(),
      commonjs(),
      typescript({
        tsconfig: './tsconfig.json',
        declaration: true,
        // Override TypeScript compiler options to match Rollup output
        compilerOptions: {
          outDir: 'dist/core',
          rootDir: 'src/core',
        },
      }),
    ],
  },
  // React entry (with styles)
  {
    input: 'src/react/index.ts',
    output: {
      dir: 'dist/react',
      format: 'es',
      exports: 'named',
      preserveModules: true,
      preserveModulesRoot: 'src/react',
      entryFileNames: '[name].mjs',
      sourcemap: true,
    },
    external: ['react', 'react-dom'], // Peer dependencies
    plugins: [
      resolve(),
      commonjs(),
      typescript({
        tsconfig: './tsconfig.json',
        declaration: true,
        // Override TypeScript compiler options to match Rollup output
        // rootDir must be 'src' to include both core and react directories
        compilerOptions: {
          outDir: 'dist/react',
          rootDir: 'src',
        },
      }),
      postcss({
        extract: 'styles.css',
        minimize: true,
      }),
      terser(),
    ],
  },
];
