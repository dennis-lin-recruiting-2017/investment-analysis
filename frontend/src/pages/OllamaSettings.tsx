import LLMSettingsForm from '../components/LLMSettingsForm';

export default function OllamaSettings() {
  return (
    <LLMSettingsForm
      provider="ollama"
      title="Ollama"
      description="Configure the Ollama host, model, and generation parameters used by the app."
    />
  );
}
