export function getSoulCoreImageUrl(creatureName: string): string {
  const filename = `${creatureName.replace(/ /g, '_')}_Soul_Core.gif`
  return `/assets/soulcores/${filename}`
}
