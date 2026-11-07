import "./channelCard.css";
import {For, Show} from "solid-js";

interface ChannelCardProps {
    channel: Channel;
    onEdit: (channel: Channel) => void;
}

export function ChannelCard(props: ChannelCardProps) {
    const handleEdit = () => props.onEdit(props.channel);

    return (
        <article class="channel-card">
            <header class="channel-card__header">
                <div class="channel-card__avatar">
                    <img
                        src={props.channel.icon_url}
                        alt={`${props.channel.name} logo`}
                        loading="lazy"
                    />
                </div>

                <div class="channel-card__identity">
                    <h3>{props.channel.name}</h3>
                    <span class="channel-card__provider">
            {props.channel.provider}
          </span>
                </div>

                <span class="channel-card__language">
          {props.channel.language.toUpperCase()}
        </span>
            </header>

            <div class="channel-card__content">
        <span class="channel-card__label">
          Assigned leagues
        </span>

                <div class="channel-card__leagues">
                    <Show
                        when={props.channel.leagues?.length}
                        fallback={
                            <span class="channel-card__empty">
                No leagues assigned
              </span>
                        }
                    >
                        <For each={props.channel.leagues}>
                            {(league) => (
                                <span class="channel-card__league">
                  {league.name}
                </span>
                            )}
                        </For>
                    </Show>
                </div>
            </div>

            <button
                type="button"
                class="channel-card__edit"
                onClick={handleEdit}
                aria-label={`Edit ${props.channel.name}`}
            >
                <span aria-hidden="true">✎</span>
                Edit channel
            </button>
        </article>
    );
}
