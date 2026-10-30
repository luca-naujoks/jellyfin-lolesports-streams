import {createSignal, For, Show} from "solid-js";
import "./channelForm.css";

interface ChannelFormProps {
    channel?: Channel;
    availableLeagues: League[];
    onSave: (channel: Channel) => Promise<void>;
    onClose: () => void;
}

export function ChannelForm(props: ChannelFormProps) {
    const [id, setId] = createSignal(props.channel?.id ?? "");
    const [name, setName] = createSignal(props.channel?.name ?? "");
    const [provider, setProvider] = createSignal(
        props.channel?.provider ?? "twitch"
    );
    const [language, setLanguage] = createSignal(
        props.channel?.language ?? "en"
    );
    const [iconUrl, setIconUrl] = createSignal(
        props.channel?.icon_url ?? ""
    );
    const [selectedLeagues, setSelectedLeagues] = createSignal<string[]>(
        props.channel?.leagues.map((league) => league.id) ?? []
    );

    const [saving, setSaving] = createSignal(false);
    const [error, setError] = createSignal("");

    const isEditing = () => props.channel !== undefined;

    function toggleLeague(leagueId: string) {
        setSelectedLeagues((current) =>
            current.includes(leagueId)
                ? current.filter((id) => id !== leagueId)
                : [...current, leagueId]
        );
    }

    async function handleSubmit(event: SubmitEvent) {
        event.preventDefault();

        if (!id().trim() || !name().trim()) {
            setError("Channel ID and name are required.");
            return;
        }

        setSaving(true);
        setError("");

        try {
            const channel: Channel = {
                id: id().trim(),
                name: name().trim(),
                provider: provider(),
                language: language(),
                icon_url: iconUrl().trim(),
                leagues: props.availableLeagues.filter((league) =>
                    selectedLeagues().includes(league.id)
                ),
            };

            await props.onSave(channel);
        } catch (err) {
            setError(
                err instanceof Error ? err.message : "Failed to save channel."
            );
        } finally {
            setSaving(false);
        }
    }

    return (
        <div
            class="channel-modal"
            onClick={(event) => {
                if (event.target === event.currentTarget) {
                    props.onClose();
                }
            }}
        >
            <div
                class="channel-modal__content"
                role="dialog"
                aria-modal="true"
                aria-labelledby="channel-form-title"
            >
                <div class="channel-modal__header">
                    <div>
                        <h2 id="channel-form-title">
                            {isEditing() ? "Edit channel" : "Create channel"}
                        </h2>
                        <p>Configure your live TV broadcaster</p>
                    </div>

                    <button
                        type="button"
                        class="channel-modal__close"
                        onClick={props.onClose}
                        aria-label="Close"
                    >
                        ×
                    </button>
                </div>

                <form onSubmit={handleSubmit} class="channel-form">
                    <div class="channel-form__row">
                        <label>
                            Channel ID
                            <input
                                value={id()}
                                onInput={(event) => setId(event.currentTarget.value)}
                                disabled={isEditing()}
                                required
                                placeholder="riotgames"
                            />
                        </label>

                        <label>
                            Display name
                            <input
                                value={name()}
                                onInput={(event) => setName(event.currentTarget.value)}
                                required
                                placeholder="Riot Games"
                            />
                        </label>
                    </div>

                    <div class="channel-form__row">
                        <label>
                            Provider
                            <select
                                value={provider()}
                                onChange={(event) =>
                                    setProvider(event.currentTarget.value)
                                }
                            >
                                <option value="twitch">Twitch</option>
                                <option value="youtube" disabled>YouTube</option>
                            </select>
                        </label>

                        <label>
                            Language
                            <select
                                value={language()}
                                onChange={(event) =>
                                    setLanguage(event.currentTarget.value)
                                }
                            >
                                <option value="en">English</option>
                                <option value="de">German</option>
                            </select>
                        </label>
                    </div>

                    <label>
                        Channel logo URL
                        <input
                            type="url"
                            value={iconUrl()}
                            onInput={(event) =>
                                setIconUrl(event.currentTarget.value)
                            }
                            placeholder="https://example.com/logo.png"
                        />
                    </label>

                    <Show when={iconUrl()}>
                        <div class="channel-form__preview">
                            <img
                                src={iconUrl()}
                                alt="Channel logo preview"
                            />
                        </div>
                    </Show>

                    <div class="channel-form__leagues">
                        <h3>Assigned leagues</h3>
                        <p>Select leagues usually broadcasted by this channel</p>

                        <div class="channel-form__league-list">
                            <For each={props.availableLeagues}>
                                {(league) => (
                                    <label class="channel-form__league">
                                        <input
                                            type="checkbox"
                                            checked={selectedLeagues().includes(league.id)}
                                            onChange={() => toggleLeague(league.id)}
                                        />
                                        <span>{league.name}</span>
                                        <span class="channel-form__league-region">
                      {league.region}
                    </span>
                                    </label>
                                )}
                            </For>
                        </div>
                    </div>

                    <Show when={error()}>
                        <p class="channel-form__error" role="alert">
                            {error()}
                        </p>
                    </Show>

                    <div class="channel-form__actions">
                        <button
                            type="button"
                            class="channel-form__cancel"
                            onClick={props.onClose}
                        >
                            Cancel
                        </button>

                        <button
                            type="submit"
                            class="channel-form__save"
                            disabled={saving()}
                        >
                            {saving()
                                ? "Saving..."
                                : isEditing()
                                    ? "Save changes"
                                    : "Create channel"}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
}
