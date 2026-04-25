import LLMSettingsForm from '../components/LLMSettingsForm';

export default function LMStudioSettings() {
  return (
    <LLMSettingsForm
      provider="lm-studio"
      title="LM Studio"
      description="Configure the local LM Studio endpoint, default model, and prompt behavior."
    />
  );
}
