# templates/

Reusable document templates. When Claude (or you) produces a new document that should match the team's house style, the shape comes from here.

## What goes here

- **`.potx`** — PowerPoint templates. Slide masters, brand colors, standard layouts.
- **`.dotx`** — Word templates. Report headers, footers, paragraph styles.
- **`.xltx`** — Excel templates. Standard sheets, named ranges, conditional formats.
- **`.md` skeletons** — markdown templates for meeting briefs, decision proposals, project kickoffs. Useful as input scaffolding before converting to Office formats.

## How Claude uses templates

Claude produces markdown drafts by default. When a draft is ready for a real audience, you convert it to Office format via one of these templates:

1. Tell Claude which template to follow (by filename).
2. Claude writes the markdown structured to match that template's sections.
3. You open the template in Word/PowerPoint/Excel and paste the rendered content.

Claude does not edit `.potx`/`.dotx`/`.xltx` files directly — they're binary, and a wrong edit corrupts every future document built from them.

## Conventions

- Use descriptive filenames: `quarterly-board-deck.potx`, not `template_v2.potx`.
- Date-prefix versions if you maintain multiple revisions: `2026-q1_board-deck.potx`.
- Note in this README which template applies to which kind of document — it'll save Claude (and new coworkers) a guess.

## Active templates

> **Replace this list as you add templates.**

- `aeven-template.potx` — standard slide deck for internal presentations.
- (add more here)
