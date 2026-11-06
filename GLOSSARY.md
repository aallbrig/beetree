# Glossary

**Agent**: a character or entity that runs a tree. In the product, an entry in a person's inventory.

**Tree**: a behavior tree described in the engine-agnostic `.beetree.yaml` format.

**Template**: a ready-made tree inspired by a common game trope, offered as a starting point.

**Building block**: one of the node types a tree is made of: Action, Condition, Sequence, Fallback, Parallel, Decorator.

**Technique**: a named way of writing trees (recursive, event-driven, graph-based, concurrent, and so on), taught in Learn.

**Step-through**: simulating a tree one leaf at a time, where the author declares each leaf's status (Success, Failure or Running) and the composites resolve from those answers. Contrast with a full evaluation, which runs the whole traversal at once.

**Communitree**: the email list people join from the home and sign-up pages.

**Release**: a tagged version with release notes and artifacts, deployed to beetreecraft.com by CI.
