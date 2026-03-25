const path = require('path');
const { ConsoleRemotePlugin } = require('@openshift-console/dynamic-plugin-sdk-webpack');

const config = {
  mode: 'development',
  entry: {},
  output: {
    path: path.resolve(__dirname, 'dist'),
    filename: '[name].js',
    chunkFilename: '[name].chunk.js',
    publicPath: '/',
  },
  resolve: {
    extensions: ['.ts', '.tsx', '.js', '.jsx'],
  },
  module: {
    rules: [
      {
        test: /\.(tsx?|jsx?)$/,
        exclude: /node_modules/,
        use: [
          {
            loader: 'ts-loader',
            options: { transpileOnly: true },
          },
        ],
      },
      {
        test: /\.css$/,
        include: /node_modules\/@patternfly\/react-topology/,
        use: ['style-loader', 'css-loader'],
      },
    ],
  },
  plugins: [
    new ConsoleRemotePlugin({
      pluginMetadata: {
        name: 'vcf-migration-console',
        version: '0.0.1',
        exposedModules: {
          migrationPlugin: path.resolve(__dirname, 'src/app/index.ts'),
        },
      },
      extensions: require('./src/console-extensions.json'),
    }),
  ],
  devtool: 'source-map',
  optimization: {
    chunkIds: 'named',
    minimize: false,
  },
};

module.exports = config;
