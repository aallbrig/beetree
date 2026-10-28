---
title: "Website manual"
description: "Create an account, build a tree, watch it run in 3D, and share it."
weight: 10
---

## 1. Create your account

1. Open [Sign up](/app/#/signup).
2. Choose a **username** (3–32 letters, digits, `_` or `-`). It is your public name on trees, comments and forum posts.
3. Enter your **email**, your **date of birth** and a **password** (at least 10 characters, with upper and lower case letters and a number).
4. Tick the box confirming you are **18 or older**. The age check is enforced on the server too, so accounts for under-18s cannot be created.
5. We email you a **confirmation code**. Enter it on the confirm page to activate your account, then sign in.

**Forgot your password?** Choose *Forgot your password?* on the sign-in page, enter your username or email, then enter the code we send and a new password.

## 2. Explore (no account needed)

[Explore](/app/#/explore) is a live 3D world. Three agents run behavior trees right in your browser:

- **Patroller** walks a loop of waypoints.
- **Guard** patrols, but chases and attacks the intruder when it gets close.
- **Coward** runs away from the intruder, and heals when it is safe.

Try this:

- **Click the ground** to move the red intruder and watch the Guard react.
- Drag the **health** slider down and see how the trees change what they do.
- Click an agent (or its name in the side panel) to see **its tree**. The coloured dot beside each node shows what it is doing right now: yellow is running, green succeeded, red failed.
- Drag to orbit the camera, scroll to zoom.

## 3. Build a tree

Go to your [Dashboard](/app/#/dashboard) and choose **New tree**. (You need to be signed in.)

The builder has a live 3D preview on the right, running your tree as you edit it.

**Design tab**

- Select a node in the outline, choose a type from the dropdown, and press **Add child**.
- Use **Remove**, **↑** and **↓** to delete and reorder.
- Edit the selected node's **name**, **implementation id** (the class or function your engine provides), **description** and **parameters** (a JSON object, such as `{"seconds": 2}`).
- Warnings appear under the *Save* button as you build, for example when a sequence has no children.

**YAML tab**: the same tree as text. Edit it and press **Apply YAML**.

**Engine code tab**: pick Unity, Unreal or Godot and press **Generate** to see the code your engine would use.

Press **Validate** to run the server's validator, then **Save tree**. Unsaved drafts of new trees are remembered in your browser.

### Make the preview act out your tree

The preview understands common game-AI names. Name your leaves with words like these and the character will do something visible:

| Words in the name | Behaviour |
|---|---|
| `can_see_target`, `has_target`, `detect` | Condition: is the intruder within about 12 metres? |
| `in_range` | Condition: is the intruder within melee reach? |
| `health_low`, `is_dying` | Condition: is health below 30? |
| `chase`, `pursue`, `move_to_target` | Run toward the intruder |
| `attack`, `strike`, `melee` | Swing at the intruder (fails if too far away) |
| `shoot`, `fire`, `ranged` | Ranged attack with a projectile |
| `flee`, `run_to_cover`, `retreat` | Sprint away |
| `patrol`, `move_to_waypoint` | Walk to the next waypoint |
| `next_waypoint` | Advance to the following waypoint |
| `wait`, `idle`, `search` | Stand still for `seconds` (default 1.2) |
| `heal`, `regen` | Recover health |

Anything else still runs (an action takes about a second), so any tree stays alive. This is only the preview. Your real implementations come from the generated engine code.

## 4. Put agents in a 3D scene

Open [3D Scenes](/app/#/scenes) to arrange several agents in one world.

1. Press **Add agent**, then give it a name, a character and a behavior tree. You can use a built-in sample or any of your own trees.
2. **Click the ground** to move the selected agent. Click an agent in the scene to select it.
3. Choose **Private** or **Public (shareable link)** and press **Save scene**.
4. Use **Watch it** to see the whole scene run in Explore. A public scene can be shared with a link like `https://beetreecraft.com/app/#/explore?scene=<id>`.

## 5. Share and collaborate

- Set a tree's visibility to **Public** and it appears in [Community](/app/#/community), where anyone can read, like and comment on it.
- Open any community tree and press **Fork in builder** to start your own copy.
- Join the [Forum](/app/#/forum) to ask for help or show off your work.
- Your [Dashboard](/app/#/dashboard) shows your trees with their likes and comments, plus your forum activity.

## 6. Your profile

On your [profile](/app/#/profile) you can change your display name and bio, and copy a short-lived **API token** for the [REST API](/docs/api/) and [MCP server](/docs/mcp/).
