import {createSignal, For, onSettled, Show,} from "solid-js";

import "./App.css";

import {ChannelCard} from "../components/channelCard";
import {CreateChannelCard} from "../components/createChannelCard";
import {ChannelForm} from "../components/channelForm";

export default function App() {
    const [channels, setChannels] = createSignal<Channel[]>([]);
    const [availableLeagues, setAvailableLeagues] =
        createSignal<League[]>([]);

    const [formOpen, setFormOpen] = createSignal(false);
    const [editingChannel, setEditingChannel] =
        createSignal<Channel | undefined>(undefined);

    async function channelFetcher(): Promise<Channel[]> {
        const response = await fetch("/api/channels");

        if (!response.ok) {
            throw new Error("Failed to fetch channels");
        }

        return await response.json();
    }

    async function loadChannels() {
        try {
            const data = await channelFetcher();
            setChannels(data);
        } catch (error) {
            console.error(error);
        }
    }

    async function loadLeagues() {
        try {
            const response = await fetch("/api/leagues");

            if (!response.ok) {
                throw new Error("Failed to fetch leagues");
            }

            setAvailableLeagues(await response.json());
        } catch (error) {
            console.error(error);
        }
    }

    onSettled(
        () => {
            void loadChannels();
            void loadLeagues();
        }
    )

    function onEdit(channel: Channel) {
        setEditingChannel(channel);
        setFormOpen(true);
    }

    function onCreate() {
        setEditingChannel(undefined);
        setFormOpen(true);
    }

    function closeForm() {
        setFormOpen(false);
        setEditingChannel(undefined);
    }

    async function saveChannel(channel: Channel) {
        const editing = editingChannel() !== undefined;

        const url = editing
            ? `/api/channels/${encodeURIComponent(channel.id)}`
            : "/api/channels";

        const response = await fetch(url, {
            method: editing ? "PUT" : "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(channel),
        });

        if (!response.ok) {
            const message = await response.text();

            throw new Error(
                message || `Failed to save channel (${response.status})`
            );
        }

        await loadChannels();
        closeForm();
    }

    return (
        <main class="app">
            <section class="channels-page">
                <h1>Jellyfin LoL Esports Bridge</h1>

                <div class="channels-page__heading">
                    <div>
                        <h2>Channels</h2>
                        <p>Manage your live TV broadcasters</p>
                    </div>

                    <span>{channels()?.length} channels</span>
                </div>

                <div class="channels-grid">
                    <For each={channels()}>
                        {(channel) => (
                            <ChannelCard
                                channel={channel}
                                onEdit={onEdit}
                            />
                        )}
                    </For>

                    <CreateChannelCard onCreate={onCreate}/>
                </div>
            </section>

            <Show when={formOpen()}>
                <ChannelForm
                    channel={editingChannel()}
                    availableLeagues={availableLeagues()}
                    onSave={saveChannel}
                    onClose={closeForm}
                />
            </Show>
        </main>
    );
}
