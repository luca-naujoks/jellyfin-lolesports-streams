import "./channelCard.css";

interface CreateChannelCardProps {
    onCreate: () => void;
}

export function CreateChannelCard(props: CreateChannelCardProps) {
    return (
        <button
            type="button"
            class="create-channel-card"
            onClick={props.onCreate}
        >
            <div class="create-channel-card__icon">
                +
            </div>

            <span class="create-channel-card__title">
        Add new channel
      </span>

            <span class="create-channel-card__description">
        Connect another live broadcaster
      </span>
        </button>
    );
}
