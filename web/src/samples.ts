/** Built-in trees so Explore works before anyone has signed in. */
export interface Sample {
  id: string;
  title: string;
  yaml: string;
}

export const SAMPLES: Sample[] = [
  {
    id: 'sample:patrol',
    title: 'Patrol loop',
    yaml: `version: "1.0"
metadata:
  name: patrol
  description: Walk the waypoints forever
tree:
  type: decorator
  name: patrol_forever
  decorator: repeat
  children:
    - type: sequence
      name: patrol_cycle
      children:
        - {type: action, name: move_to_waypoint}
        - {type: action, name: wait_at_waypoint}
        - {type: action, name: next_waypoint}
`,
  },
  {
    id: 'sample:guard',
    title: 'Guard (chase and fight)',
    yaml: `version: "1.0"
metadata:
  name: guard
  description: Patrol, spot intruders, chase them down and attack
tree:
  type: selector
  name: guard_ai
  children:
    - type: sequence
      name: combat
      children:
        - {type: condition, name: can_see_target}
        - type: selector
          name: attack_or_close
          children:
            - type: sequence
              name: melee
              children:
                - {type: condition, name: in_range}
                - {type: action, name: melee_attack}
            - {type: action, name: chase_target}
    - type: sequence
      name: patrol
      children:
        - {type: action, name: move_to_waypoint}
        - {type: action, name: wait_at_waypoint}
        - {type: action, name: next_waypoint}
`,
  },
  {
    id: 'sample:coward',
    title: 'Coward (flee when hurt)',
    yaml: `version: "1.0"
metadata:
  name: coward
  description: Runs from intruders and recovers when safe
tree:
  type: selector
  name: coward_ai
  children:
    - type: sequence
      name: flee
      children:
        - {type: condition, name: can_see_target}
        - {type: action, name: run_to_cover}
    - type: sequence
      name: recover
      children:
        - {type: action, name: heal}
        - {type: action, name: wait_and_rest}
`,
  },
  {
    id: 'sample:boss',
    title: 'Boss with phases',
    yaml: `version: "1.0"
metadata:
  name: boss
  description: Ranged opener, melee up close, retreat and heal when hurt
tree:
  type: selector
  name: boss_ai
  children:
    - type: sequence
      name: desperate_phase
      children:
        - {type: condition, name: health_is_low}
        - {type: action, name: retreat_to_cover}
        - {type: action, name: heal_self}
    - type: sequence
      name: melee_phase
      children:
        - {type: condition, name: target_in_range}
        - {type: action, name: melee_attack}
    - type: sequence
      name: ranged_phase
      children:
        - {type: condition, name: can_see_target}
        - {type: action, name: shoot_projectile}
    - {type: action, name: look_around}
`,
  },
  {
    id: 'sample:companion',
    title: 'Loyal companion',
    yaml: `version: "1.0"
metadata:
  name: companion
  description: Defend the owner when a threat appears, otherwise stay close
tree:
  type: selector
  name: companion_ai
  children:
    - type: sequence
      name: defend_owner
      children:
        - {type: condition, name: can_see_target}
        - type: selector
          name: engage
          children:
            - type: sequence
              name: strike
              children:
                - {type: condition, name: target_in_range}
                - {type: action, name: melee_attack}
            - {type: action, name: chase_target}
    - {type: action, name: idle_near_owner}
`,
  },
];

export const sampleById = (id: string): Sample | undefined => SAMPLES.find((s) => s.id === id);
