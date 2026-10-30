
import "./channelCard.css";

interface ChannelCardProps {
    channel: Channel;
    onEdit: (channel: Channel) => void;
}

export function ChannelCard(props: ChannelCardProps) {
    return (
        <article class="channel-card">
            <div class="channel-card__header">
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
            </div>

            <div class="channel-card__content">
        <span class="channel-card__label">
          ASSIGNED LEAGUES
        </span>

                <div class="channel-card__leagues">
                    {props.channel.leagues.length > 0 ? (
                        props.channel.leagues.map((league) => (
                            <span class="channel-card__league">
                {league.name}
              </span>
                        ))
                    ) : (
                        <span class="channel-card__empty">
              No leagues assigned
            </span>
                    )}
                </div>
            </div>

            <button
                type="button"
                class="channel-card__edit"
                onClick={() => props.onEdit(props.channel)}
            >
                <span>✎</span>
                Edit channel
            </button>
        </article>
    );
}
