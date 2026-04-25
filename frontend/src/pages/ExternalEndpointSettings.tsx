import LLMSettingsForm from '../components/LLMSettingsForm';

export default function ExternalEndpointSettings() {
  return (
    <LLMSettingsForm
      provider="external-endpoint"
      title="External Endpoint"
      description="Configure a hosted API endpoint, authentication key, and default model settings."
    />
  );
}
