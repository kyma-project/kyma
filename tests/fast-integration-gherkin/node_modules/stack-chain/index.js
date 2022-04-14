// If a another copy (same version or not) of stack-chain exists it will result
// in wrong stack traces (most likely dublicate callSites).
if (global._stackChain) {
  // In case the version match, we can simply return the first initialized copy
  if (global._stackChain.version === require('./package.json').version) {
    module.exports = global._stackChain;
  }
  // The version don't match, this is really bad. Lets just throw
  else {
    throw new Error('Conflicting version of stack-chain found');
  }
}
// Yay, no other stack-chain copy exists, yet :/
else {
  module.exports = global._stackChain = require('./stack-chain');
}
