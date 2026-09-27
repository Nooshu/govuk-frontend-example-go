import { runBuildStylesCli } from './build-styles.mjs';

runBuildStylesCli().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
