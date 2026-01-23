<script>
  import { Route } from "yrv";
  import Credits from "./Credits.svelte";
  import Welcome from "./Welcome.svelte";
  import CreatorLayout from "./creator/Creator.svelte";
  import RoomsEditor from "./creator/RoomsEditor.svelte";
  import ItemsEditor from "./creator/ItemsEditor.svelte";
  import ItemTemplatesEditor from "./creator/ItemTemplatesEditor.svelte";
  import CharacterTemplatesEditor from "./creator/CharacterTemplatesEditor.svelte";
  import NPCsEditor from "./creator/NPCsEditor.svelte";
  import DialogsEditor from "./creator/DialogsEditor.svelte";
  import DialogGraphEditor from "./creator/DialogGraphEditor.svelte";
  import ScriptsEditor from "./creator/ScriptsEditor.svelte";
  import WorldEditor from "./creator/GridWorldEditor.svelte";
  import Characters from "./characters/Characters.svelte";
  import NewCharacter from "./characters/NewCharacter.svelte";
  import UserForm from "./UserForm.svelte";
  import News from "./news/News.svelte";
</script>

<main class="min-h-[calc(100vh-72px)]">
  <Route exact path="/account" component="{UserForm}" />
  <!-- yrv only matches Route components that are direct children of Router, so we
       define creator sub-routes here and use Creator as a shared layout shell. -->
  <Route exact path="/creator">
    <CreatorLayout>
      <RoomsEditor />
    </CreatorLayout>
  </Route>
  <Route exact path="/creator/rooms">
    <CreatorLayout>
      <RoomsEditor />
    </CreatorLayout>
  </Route>
  <Route exact path="/creator/items">
    <CreatorLayout>
      <ItemsEditor />
    </CreatorLayout>
  </Route>
  <Route exact path="/creator/item-templates">
    <CreatorLayout>
      <ItemTemplatesEditor />
    </CreatorLayout>
  </Route>
  <Route exact path="/creator/character-templates">
    <CreatorLayout>
      <CharacterTemplatesEditor />
    </CreatorLayout>
  </Route>
  <Route exact path="/creator/npcs">
    <CreatorLayout>
      <NPCsEditor />
    </CreatorLayout>
  </Route>
  <Route exact path="/creator/dialogs">
    <CreatorLayout>
      <DialogsEditor />
    </CreatorLayout>
  </Route>
  <Route exact path="/creator/dialog-graph">
    <CreatorLayout>
      <DialogGraphEditor />
    </CreatorLayout>
  </Route>
  <Route exact path="/creator/scripts">
    <CreatorLayout>
      <ScriptsEditor />
    </CreatorLayout>
  </Route>
  <Route exact path="/creator/world">
    <CreatorLayout>
      <WorldEditor />
    </CreatorLayout>
  </Route>
  <Route exact path="/list" component="{Characters}" />
  <Route exact path="/characters/new" component="{NewCharacter}" />
  <Route exact path="/credits" component="{Credits}" />
  <Route exact path="/news" component="{News}" />

  <Route exact path="/">
    <Welcome />
  </Route>
</main>
