import commander from '../../../examples/commander.beetree.yaml?raw';
import shieldbearer from '../../../examples/shieldbearer.beetree.yaml?raw';
import skirmisher from '../../../examples/skirmisher.beetree.yaml?raw';
import trooper from '../../../examples/trooper.beetree.yaml?raw';

/** The four Explore scene 2 trees, straight from the example files so docs and scene cannot drift. */
export const COMBAT_TREES: Record<string, string> = {
  'example:trooper': trooper,
  'example:skirmisher': skirmisher,
  'example:shieldbearer': shieldbearer,
  'example:commander': commander,
};
